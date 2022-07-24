package handle

import (
	"fmt"

	"github.com/geeklubcn/richman/service"
	"github.com/gin-gonic/gin"
	larkCore "github.com/larksuite/oapi-sdk-go/core"
	"github.com/larksuite/oapi-sdk-go/event"
	"github.com/sirupsen/logrus"
)

type Feishu interface {
	Webhook(ctx *gin.Context)
}

type feishu struct {
	appSvc  service.AppSvc
	bookSvc service.BookSvc
}

func NewFeishu(appSvc service.AppSvc, bookSvc service.BookSvc) Feishu {
	f := &feishu{
		appSvc:  appSvc,
		bookSvc: bookSvc,
	}
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
	appId := ctx.Param("app_id")
	facade, exist := facades[appId]
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
