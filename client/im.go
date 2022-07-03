package client

import (
	"context"
	"encoding/json"

	"github.com/larksuite/oapi-sdk-go/core"
	"github.com/larksuite/oapi-sdk-go/core/config"
	"github.com/larksuite/oapi-sdk-go/core/tools"
	oapi_im "github.com/larksuite/oapi-sdk-go/service/im/v1"
	"github.com/sirupsen/logrus"
)

const (
	MsgTypeText = "text"
)

type Im interface {
	ReplyTextMsg(ctx context.Context, messageID, content string) (string, error)
}

type feishuIm struct {
	s *oapi_im.Service
}

func NewFeishuIm(cfg *config.Config) Im {
	return &feishuIm{s: oapi_im.NewService(cfg)}
}

func (i *feishuIm) ReplyTextMsg(ctx context.Context, messageID, content string) (string, error) {
	oc := core.WrapContext(ctx)
	msg := TextMsg{Text: content}
	msgC, _ := json.Marshal(msg)
	reqCall := i.s.Messages.Reply(oc, &oapi_im.MessageReplyReqBody{
		MsgType: MsgTypeText,
		Content: string(msgC),
	})
	reqCall.SetMessageId(messageID)
	message, err := reqCall.Do()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("ReplyTextMsg fail! messageID:%s,content:%s,response:%s", messageID, content, tools.Prettify(message))
		return "", err
	}
	logrus.WithContext(ctx).Debugf("ReplyTextMsg messageID:%s,content:%s,response:%s", messageID, content, tools.Prettify(message))
	return message.MessageId, nil
}

// TextMsg 文本消息
type TextMsg struct {
	Text string `json:"text"`
}
