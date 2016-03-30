package libcentrifugo

import (
	"encoding/json"
	"testing"

	//"github.com/mailru/easyjson"
)

func BenchmarkCommandUnmarshal(b *testing.B) {
	cmd := ApiCommand{
		Method: "publish",
		Params: []byte("{\"channel\":\"test\", \"data\": {}}"),
	}
	msg, _ := json.Marshal(cmd)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var c ApiCommand
		err := json.Unmarshal(msg, &c)
		if err != nil {
			panic(err)
		}
	}
}
