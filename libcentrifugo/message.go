package libcentrifugo

import (
	"strconv"
	"time"

	"github.com/centrifugal/centrifugo/libcentrifugo/engine"
	"github.com/centrifugal/centrifugo/libcentrifugo/raw"
	"github.com/nats-io/nuid"
)

func newClientInfo(user UserID, client ConnID, defaultInfo *raw.Raw, channelInfo *raw.Raw) engine.ClientInfo {
	return engine.ClientInfo{
		User:        string(user),
		Client:      string(client),
		DefaultInfo: defaultInfo,
		ChannelInfo: channelInfo,
	}
}

func newMessage(ch Channel, data []byte, client ConnID, info *ClientInfo) *engine.Message {
	raw := raw.Raw(data)
	return &engine.Message{
		UID:       nuid.Next(),
		Timestamp: strconv.FormatInt(time.Now().Unix(), 10),
		Info:      info,
		Channel:   string(ch),
		Data:      &raw,
		Client:    string(client),
	}
}

func newJoinMessage(ch Channel, info ClientInfo) *engine.JoinMessage {
	return &engine.JoinMessage{
		Channel: string(ch),
		Data:    info,
	}
}

func newLeaveMessage(ch Channel, info ClientInfo) *engine.LeaveMessage {
	return &engine.LeaveMessage{
		Channel: string(ch),
		Data:    info,
	}
}

func newAdminMessage(uid string, method string, params []byte) *engine.AdminMessage {
	raw := raw.Raw(data)
	return &engine.AdminMessage{
		UID:    uid,
		Method: method,
		Params: &raw,
	}
}

func newControlMessage(uid string, method string, params []byte) *engine.ControlMessage {
	raw := raw.Raw(data)
	return &engine.ControlMessage{
		UID:    uid,
		Method: method,
		Params: &raw,
	}
}
