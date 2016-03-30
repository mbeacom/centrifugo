package libcentrifugo

import (
	"encoding/json"
	"testing"

	"github.com/mailru/easyjson"
)

func BenchmarkCommandUnmarshal(b *testing.B) {
	cmd := apiCommand{
		Method: "publish",
		Params: []byte("{\"channel\":\"test\", \"data\": {}}"),
	}
	msg, _ := json.Marshal(cmd)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var c apiCommand
		err := easyjson.Unmarshal(msg, &c)
		if err != nil {
			panic(err)
		}
	}
}
