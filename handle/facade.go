package handle

import (
	"encoding/json"
	"fmt"
	"github.com/geeklubcn/richman/client"
	"github.com/geeklubcn/richman/model"
	"github.com/geeklubcn/richman/service"
	"github.com/hashicorp/golang-lru"
	larkCore "github.com/larksuite/oapi-sdk-go/core"
	"github.com/larksuite/oapi-sdk-go/core/config"
	"github.com/larksuite/oapi-sdk-go/core/tools"
	larkIm "github.com/larksuite/oapi-sdk-go/service/im/v1"
	"github.com/sirupsen/logrus"
)

var facades = map[string]*Facade{}
var idempotent *lru.Cache

type Facade struct {
	Conf      *config.Config
	appSvc    service.AppSvc
	authorSvc service.AuthorSvc

	BillSvc service.BillSvc
	Ims     client.Im
}

func Init(appSvc service.AppSvc, authorSvc service.AuthorSvc, bookSvc service.BookSvc) {
	idempotent, _ = lru.New(256)
	for _, a := range appSvc.FindAll() {
		if a.AppId != "" {
			if _, exist := facades[a.AppId]; !exist {
				_ = register(*a, authorSvc, bookSvc)
			}
		}
	}
}

func register(app model.App, authorSvc service.AuthorSvc, bookSvc service.BookSvc) error {
	appSettings := larkCore.NewInternalAppSettings(
		larkCore.SetAppCredentials(app.AppId, app.AppSecret),
		larkCore.SetAppEventKey(app.VerificationToken, ""),
	)
	conf := larkCore.NewConfig(larkCore.DomainFeiShu, appSettings)

	facades[app.AppId] = &Facade{
		Conf:      conf,
		authorSvc: authorSvc,
		BillSvc:   service.NewBillSvc(app.AppId, app.AppSecret, bookSvc),
		Ims:       client.NewFeishuIm(conf),
	}
	larkIm.SetMessageReceiveEventHandler(conf, imMessageReceiveV1)

	return nil
}

func imMessageReceiveV1(ctx *larkCore.Context, event *larkIm.MessageReceiveEvent) error {
	defer func() {
		if p := recover(); p != nil {
			logrus.WithContext(ctx).Errorf("handle message receive error. messageId: %s err: %s", event.Event.Message.MessageId, p)
		}
	}()
	logrus.WithContext(ctx).Infof("receive event:%+v", tools.Prettify(event))
	if idempotent.Contains(event.Header.EventID) {
		logrus.WithContext(ctx).Infof("ignore repeat event:%+v", tools.Prettify(event))
		return nil
	}
	idempotent.Add(event.Header.EventID, true)

	switch event.Event.Message.MessageType {
	case client.MsgTypeText:
		var msg client.TextMsg
		err := json.Unmarshal([]byte(event.Event.Message.Content), &msg)
		if err != nil {
			return err
		}
		facade, _ := facades[event.Header.AppID]

		resMsg := facade.BillSvc.Record(event.Header.AppID, event.Event.Sender.SenderId.OpenId, msg.Text, model.CategoryFeishu)
		mid, err := facade.Ims.ReplyTextMsg(ctx, event.Event.Message.MessageId, resMsg)
		if err != nil {
			return err
		}
		logrus.WithContext(ctx).Infof("reply message %s -> %s", event.Event.Message.MessageId, mid)
		return nil
	default:
		return fmt.Errorf("unsupprt msg! type:%s", event.Event.Message.MessageType)
	}
}
