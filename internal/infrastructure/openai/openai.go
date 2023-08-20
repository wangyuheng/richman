package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/domain"
	"net/http"
)

type openAICaller struct {
	url string
	key string
}

func NewOpenAICaller(cfg *config.Config) domain.AIService {
	return &openAICaller{url: cfg.AiURL, key: cfg.AiKey}
}

func (o *openAICaller) CallFunctions(content string, ai domain.AI) (*domain.AIMessage, error) {
	ctx := context.Background()
	data := domain.AIReq{
		Model: "gpt-3.5-turbo-0613",
		Messages: []domain.AIMessage{
			{
				Role:    "system",
				Content: ai.Introduction,
			},
			{
				Role:    "user",
				Content: content,
			},
		},
		Functions: ai.Functions,
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
	var response domain.AIResp
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
