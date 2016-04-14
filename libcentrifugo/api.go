package libcentrifugo

import (
	"encoding/json"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/centrifugal/centrifugo/Godeps/_workspace/src/github.com/FZambia/go-logger"
	"github.com/mailru/easyjson"
)

// apiCmd builds API command and dispatches it into correct handler method.
func (app *Application) apiCmd(command apiCommand) (*response, error) {

	var err error
	var resp *response

	method := command.Method
	params := command.Params

	switch method {
	case "publish":
		var cmd publishAPICommand

		channelBytes, err := jsonparser.GetString(params, "channel")
		if err != nil {
			logger.ERROR.Println(err)
			return nil, ErrInvalidMessage
		}

		dataBytes, vType, _, err := jsonparser.Get(params, "data")

		if err != nil {
			logger.ERROR.Println(err)
			return nil, ErrInvalidMessage
		}

		clientBytes, err := jsonparser.GetString(params, "client")
		if err != nil {
			if err == jsonparser.KeyPathNotFoundError {
				clientBytes = ""
			} else {
				logger.ERROR.Println(err)
				return nil, ErrInvalidMessage
			}
		}

		if vType == jsonparser.Null {
			cmd.Data = json.RawMessage([]byte("null"))
		} else if vType == jsonparser.String {
			cmd.Data = json.RawMessage([]byte(fmt.Sprintf("\"%s\"", string(dataBytes))))
		} else {
			cmd.Data = json.RawMessage(dataBytes)
		}
		cmd.Client = ConnID(clientBytes)
		cmd.Channel = Channel(channelBytes)

		resp, err = app.publishCmd(&cmd)
	case "broadcast":
		var cmd broadcastAPICommand

		var channels []Channel

		channelsBytes, _, _, err := jsonparser.Get(params, "channels")
		if err != nil {
			logger.ERROR.Println(err)
			return nil, ErrInvalidMessage
		}

		// You can use `ArrayEach` helper to iterate items
		jsonparser.ArrayEach(channelsBytes, func(value []byte, vType jsonparser.ValueType, offset int, err error) {
			if vType != jsonparser.String {
				return
			}
			channels = append(channels, Channel(value))
		})

		dataBytes, vType, _, err := jsonparser.Get(params, "data")
		if err != nil {
			logger.ERROR.Println(err)
			return nil, ErrInvalidMessage
		}

		clientBytes, err := jsonparser.GetString(params, "client")
		if err != nil {
			if err == jsonparser.KeyPathNotFoundError {
				clientBytes = ""
			} else {
				logger.ERROR.Println(err)
				return nil, ErrInvalidMessage
			}
		}

		if vType == jsonparser.Null {
			cmd.Data = json.RawMessage([]byte("null"))
		} else if vType == jsonparser.String {
			cmd.Data = json.RawMessage([]byte(fmt.Sprintf("\"%s\"", string(dataBytes))))
		} else {
			cmd.Data = json.RawMessage(dataBytes)
		}
		cmd.Client = ConnID(clientBytes)
		cmd.Channels = channels

		resp, err = app.broadcastCmd(&cmd)
	case "unsubscribe":
		var cmd unsubscribeAPICommand
		err = easyjson.Unmarshal(params, &cmd)
		if err != nil {
			logger.ERROR.Println(err)
			return nil, ErrInvalidMessage
		}
		resp, err = app.unsubcribeCmd(&cmd)
	case "disconnect":
		var cmd disconnectAPICommand
		err = easyjson.Unmarshal(params, &cmd)
		if err != nil {
			logger.ERROR.Println(err)
			return nil, ErrInvalidMessage
		}
		resp, err = app.disconnectCmd(&cmd)
	case "presence":
		var cmd presenceAPICommand
		err = easyjson.Unmarshal(params, &cmd)
		if err != nil {
			logger.ERROR.Println(err)
			return nil, ErrInvalidMessage
		}
		resp, err = app.presenceCmd(&cmd)
	case "history":
		var cmd historyAPICommand
		err = easyjson.Unmarshal(params, &cmd)
		if err != nil {
			logger.ERROR.Println(err)
			return nil, ErrInvalidMessage
		}
		resp, err = app.historyCmd(&cmd)
	case "channels":
		resp, err = app.channelsCmd()
	case "stats":
		resp, err = app.statsCmd()
	case "node":
		resp, err = app.nodeCmd()
	default:
		return nil, ErrMethodNotFound
	}
	if err != nil {
		return nil, err
	}

	resp.UID = command.UID

	return resp, nil
}

// publishCmd publishes data into channel.
func (app *Application) publishCmd(cmd *publishAPICommand) (*response, error) {
	resp := newResponse("publish")
	channel := cmd.Channel
	data := cmd.Data
	err := app.publish(channel, data, cmd.Client, nil, false)
	if err != nil {
		resp.Err(err)
		return resp, nil
	}
	return resp, nil
}

// broadcastCmd publishes data into multiple channels.
func (app *Application) broadcastCmd(cmd *broadcastAPICommand) (*response, error) {
	resp := newResponse("broadcast")
	channels := cmd.Channels
	data := cmd.Data
	if len(channels) == 0 {
		logger.ERROR.Println("channels required for broadcast")
		resp.Err(ErrInvalidMessage)
		return resp, nil
	}
	errs := make([]<-chan error, len(channels))
	for i, channel := range channels {
		errs[i] = app.publishAsync(channel, data, cmd.Client, nil, false)
	}
	var firstErr error
	for i := range errs {
		err := <-errs[i]
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			logger.ERROR.Printf("Error publishing into channel %s: %v", string(channels[i]), err.Error())
		}
	}
	if firstErr != nil {
		resp.Err(firstErr)
	}
	return resp, nil
}

// unsubscribeCmd unsubscribes project's user from channel and sends
// unsubscribe control message to other nodes.
func (app *Application) unsubcribeCmd(cmd *unsubscribeAPICommand) (*response, error) {
	resp := newResponse("unsubscribe")
	channel := cmd.Channel
	user := cmd.User
	err := app.Unsubscribe(user, channel)
	if err != nil {
		resp.Err(err)
		return resp, nil
	}
	return resp, nil
}

// disconnectCmd disconnects user by its ID and sends disconnect
// control message to other nodes so they could also disconnect this user.
func (app *Application) disconnectCmd(cmd *disconnectAPICommand) (*response, error) {
	resp := newResponse("disconnect")
	user := cmd.User
	err := app.Disconnect(user)
	if err != nil {
		resp.Err(err)
		return resp, nil
	}
	return resp, nil
}

// presenceCmd returns response with presense information for channel.
func (app *Application) presenceCmd(cmd *presenceAPICommand) (*response, error) {
	resp := newResponse("presence")
	channel := cmd.Channel
	body := &PresenceBody{
		Channel: channel,
	}
	resp.Body = body
	presence, err := app.Presence(channel)
	if err != nil {
		resp.Err(err)
		return resp, nil
	}
	body.Data = presence
	return resp, nil
}

// historyCmd returns response with history information for channel.
func (app *Application) historyCmd(cmd *historyAPICommand) (*response, error) {
	resp := newResponse("history")
	channel := cmd.Channel
	body := &HistoryBody{
		Channel: channel,
	}
	resp.Body = body
	history, err := app.History(channel)
	if err != nil {
		resp.Err(err)
		return resp, nil
	}
	body.Data = history
	return resp, nil
}

// channelsCmd returns active channels.
func (app *Application) channelsCmd() (*response, error) {
	resp := newResponse("channels")
	body := &ChannelsBody{}
	resp.Body = body
	channels, err := app.channels()
	if err != nil {
		logger.ERROR.Println(err)
		resp.Err(ErrInternalServerError)
		return resp, nil
	}
	body.Data = channels
	return resp, nil
}

// statsCmd returns active node stats.
func (app *Application) statsCmd() (*response, error) {
	resp := newResponse("stats")
	body := &StatsBody{}
	body.Data = app.stats()
	resp.Body = body
	return resp, nil
}

// nodeCmd returns simple counter metrics which update in real time for the current node only.
func (app *Application) nodeCmd() (*response, error) {
	resp := newResponse("node")
	body := &NodeBody{}
	body.Data = app.node()
	resp.Body = body
	return resp, nil
}
