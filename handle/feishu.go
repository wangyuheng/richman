package handle

import (
	"encoding/json"
	"fmt"

	"github.com/geeklubcn/richman/biz"
	"github.com/geeklubcn/richman/client"
	"github.com/geeklubcn/richman/model"
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
	Register(ctx *gin.Context)
}

type feishu struct {
	apps biz.App
	mf   map[string]*Facade
}

type Facade struct {
	Conf  *config.Config
	Bills biz.BillBiz
	Ims   client.Im
}

func NewFeishu(appId, appSecret string) Feishu {
	f := &feishu{
		apps: biz.NewApp(appId, appSecret),
		mf:   map[string]*Facade{},
	}

	for _, a := range f.apps.FindAll() {
		if _, exist := f.mf[a.AppId]; !exist {
			appSettings := larkCore.NewInternalAppSettings(
				larkCore.SetAppCredentials(a.AppId, a.AppSecret),
				larkCore.SetAppEventKey(a.VerificationToken, ""),
			)
			conf := larkCore.NewConfig(larkCore.DomainFeiShu, appSettings)

			f.mf[a.AppId] = &Facade{
				Conf:  conf,
				Bills: biz.NewBill(appId, appSecret),
				Ims:   client.NewFeishuIm(conf),
			}
			larkIm.SetMessageReceiveEventHandler(conf, f.imMessageReceiveV1)
		}
	}

	return f
}

func (f *feishu) Register(ctx *gin.Context) {
	var app model.App
	err := ctx.ShouldBind(&app)
	if err != nil {
		_ = ctx.AbortWithError(500, err)
		return
	}
	_, err = f.apps.Save(&app)
	if err != nil {
		_ = ctx.AbortWithError(500, err)
		return
	}
	appSettings := larkCore.NewInternalAppSettings(
		larkCore.SetAppCredentials(app.AppId, app.AppSecret),
		larkCore.SetAppEventKey(app.VerificationToken, ""),
	)
	conf := larkCore.NewConfig(larkCore.DomainFeiShu, appSettings)

	f.mf[app.AppId] = &Facade{
		Conf:  conf,
		Bills: biz.NewBill(app.AppId, app.AppSecret),
		Ims:   client.NewFeishuIm(conf),
	}
	larkIm.SetMessageReceiveEventHandler(conf, f.imMessageReceiveV1)

	if err != nil {
		_ = ctx.AbortWithError(500, err)
		return
	}
	ctx.AbortWithStatus(200)
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
	appId := ctx.Param("app_id")
	facade, exist := f.mf[appId]
	if !exist {
		logrus.WithContext(ctx).Errorf("appId:%s not register", appId)
		_ = ctx.AbortWithError(500, fmt.Errorf("appId:%s not register", appId))
		return
	}

	err = event.Handle(facade.Conf, req).WriteTo(ctx.Writer)
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
		facade, _ := f.mf[event.Header.AppID]

		resMsg := facade.Bills.Record(event.Event.Sender.SenderId.OpenId, msg.Text)
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
