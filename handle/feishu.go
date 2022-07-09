package handle

import (
	"encoding/json"
	"fmt"

	"github.com/geeklubcn/richman/biz"
	"github.com/geeklubcn/richman/client"
	"github.com/gin-gonic/gin"
	larkCore "github.com/larksuite/oapi-sdk-go/core"
	"github.com/larksuite/oapi-sdk-go/core/config"
	"github.com/larksuite/oapi-sdk-go/core/tools"
	"github.com/larksuite/oapi-sdk-go/event"
	larkIm "github.com/larksuite/oapi-sdk-go/service/im/v1"
	"github.com/sirupsen/logrus"
)

type Feishu interface {
	Webhook(ctx *gin.Context)
}

type feishu struct {
	conf  *config.Config
	ims   client.Im
	bills biz.BillBiz
}

func NewFeishu(appId, appSecret, verificationToken string) Feishu {
	appSettings := larkCore.NewInternalAppSettings(
		larkCore.SetAppCredentials(appId, appSecret),
		larkCore.SetAppEventKey(verificationToken, ""),
	)
	conf := larkCore.NewConfig(larkCore.DomainFeiShu, appSettings)
	f := &feishu{
		conf:  conf,
		ims:   client.NewFeishuIm(conf),
		bills: biz.NewBill(appId, appSecret),
	}

	larkIm.SetMessageReceiveEventHandler(conf, f.imMessageReceiveV1)
	return f
}

func (f *feishu) Webhook(ctx *gin.Context) {
	req, err := larkCore.ToOapiRequest(ctx.Request)
	if err != nil {
		err = larkCore.NewOapiResponseOfErr(err).WriteTo(ctx.Writer)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Errorf("parse event request error! req:%v", req)
		}
		return
	}
	err = event.Handle(f.conf, req).WriteTo(ctx.Writer)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("handle event error! req:%v", req)
	}
}

func (f *feishu) imMessageReceiveV1(ctx *larkCore.Context, event *larkIm.MessageReceiveEvent) error {
	logrus.WithContext(ctx).Infof("receive event:%+v", tools.Prettify(event))
	switch event.Event.Message.MessageType {
	case client.MsgTypeText:
		var msg client.TextMsg
		err := json.Unmarshal([]byte(event.Event.Message.Content), &msg)
		if err != nil {
			return err
		}
		resMsg := f.bills.Record(event.Event.Sender.SenderId.OpenId, msg.Text)
		mid, err := f.ims.ReplyTextMsg(ctx, event.Event.Message.MessageId, resMsg)
		if err != nil {
			return err
		}
		logrus.WithContext(ctx).Infof("reply message %s -> %s", event.Event.Message.MessageId, mid)
		return nil
	default:
		return fmt.Errorf("unsupprt msg! type:%s", event.Event.Message.MessageType)
	}
}
