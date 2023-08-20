package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/common"
	"github.com/wangyuheng/richman/internal/domain"
	"io"
	"net/http"
)

type openAIService struct {
	auditLogger domain.AuditLogService
	url         string
	key         string
}

func NewOpenAIService(cfg *config.Config, auditLogger domain.AuditLogService) domain.AIService {
	return &openAIService{
		auditLogger: auditLogger,
		url:         cfg.AiURL,
		key:         cfg.AiKey,
	}
}

func (o *openAIService) CallFunctions(ctx context.Context, content string, ai domain.AI) (*domain.AIMessage, error) {
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

	var auditResp string
	defer func() {
		o.auditLogger.Send(domain.AuditLog{
			Req:      string(payload),
			Resp:     auditResp,
			Operator: common.GetCurrentUserID(ctx),
			Key:      "openAIService_CallFunctions",
		})
	}()

	req, _ := http.NewRequest("POST", o.url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("call openai err! req:%+v, resp:%+v", req, resp)
		return nil, err
	}
	var response domain.AIResp
	respBody, _ := io.ReadAll(resp.Body)
	auditResp = string(respBody)
	err = json.Unmarshal(respBody, &response)
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
