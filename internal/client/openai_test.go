package client

import (
	"context"
	"encoding/json"
	"github.com/magiconair/properties/assert"
	"github.com/wangyuheng/richman/config"
	"testing"
)

func TestCallOpenAIFunctions(t *testing.T) {
	caller := NewOpenAICaller(&config.Config{
		AIConfig: config.AIConfig{
			AiURL: "https://api.openai.com/v1/chat/completions",
			AiKey: "xxx",
		},
	})
	msg, _ := caller.CallFunctions(context.Background(), "打车15块", []OpenAIFunction{
		{
			Name:        "bookkeeping",
			Description: "记账工具，支持记录收入支出",
			Parameters: OpenAIParameter{
				Type: "object",
				Properties: map[string]OpenAIProperty{
					"amount": {
						Type:        "string",
						Description: "账单金额",
					},
				},
			},
		},
	})
	assert.Equal(t, msg.FunctionCall.Name, "bookkeeping")
	var args map[string]string
	_ = json.Unmarshal([]byte(msg.FunctionCall.Arguments), &args)
	assert.Equal(t, args["amount"], "15")
}
