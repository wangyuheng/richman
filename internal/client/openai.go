package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/wangyuheng/richman/config"
	"net/http"
)

type OpenAICaller interface {
	CallFunctions(ctx context.Context, content string, functions []OpenAIFunction) (*OpenAIMessage, error)
}

type openAICaller struct {
	url string
	key string
}

func NewOpenAICaller(cfg *config.Config) OpenAICaller {
	return &openAICaller{url: cfg.AiURL, key: cfg.AiKey}
}

func (o *openAICaller) CallFunctions(ctx context.Context, content string, functions []OpenAIFunction) (*OpenAIMessage, error) {
	data := OpenAIReq{
		Model: "gpt-3.5-turbo-0613",
		Messages: []OpenAIMessage{
			{
				Role:    "user",
				Content: content,
			},
		},
		Functions: functions,
	}
	payload, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", o.url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("call openai err! req:%+v, resp:%+v", req, resp)
		return nil, err
	}
	var response OpenAIResp
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		logrus.WithError(err).Errorf("call openai response not json err! req:%+v, resp:%+v", req, resp)
		return nil, err
	}
	if len(response.Choices) == 0 {
		logrus.WithError(err).Errorf("call openai response not json err! req:%+v, resp:%+v", req, resp)
		return nil, fmt.Errorf("openai response choices is empty")
	}
	return &response.Choices[0].Message, nil
}

type OpenAIReq struct {
	Model     string           `json:"model"`
	Messages  []OpenAIMessage  `json:"messages"`
	Functions []OpenAIFunction `json:"functions"`
}

type OpenAIFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  OpenAIParameter `json:"parameters"`
}

type OpenAIParameter struct {
	Type       string                    `json:"type"`
	Properties map[string]OpenAIProperty `json:"properties"`
	Required   *[]string                 `json:"required,omitempty"`
}

type OpenAIProperty struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Enum        *[]string `json:"enum,omitempty"`
}

type OpenAIChoices struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

type OpenAIResp struct {
	Id      string          `json:"id"`
	Object  string          `json:"object"`
	Created int             `json:"created"`
	Model   string          `json:"model"`
	Choices []OpenAIChoices `json:"choices"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type OpenAIMessage struct {
	Role         string              `json:"role"`
	Content      string              `json:"content"`
	FunctionCall *OpenAIFunctionCall `json:"function_call,omitempty"`
}

type OpenAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}
