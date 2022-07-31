package main

import (
	"log"

	"github.com/geeklubcn/richman/mw"

	"github.com/geeklubcn/richman/service"

	"github.com/geeklubcn/richman/config"
	"github.com/geeklubcn/richman/handle"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := config.Load()
	logrus.SetLevel(cfg.LogLevel)

	logrus.Debugf("load config. %+v", cfg)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(mw.Access())

	register(r)
	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}

func register(r *gin.Engine) {
	cfg := config.GetConfig()
	appSvc := service.NewAppSvc(cfg.DbAppId, cfg.DbAppSecret)
	authorSvc := service.NewAuthorSvc(cfg.DbAppId, cfg.DbAppSecret)
	bookSvc := service.NewBookSvc(cfg.DbAppId, cfg.DbAppSecret)

	handle.Init(appSvc, authorSvc, bookSvc)

	feishu := handle.NewFeishu(appSvc, bookSvc)
	f := r.Group("/feishu")
	{
		f.POST("/webhook/:app_id", feishu.Webhook)
	}

	admin := handle.NewAdmin(appSvc, authorSvc, bookSvc)
	a := r.Group("/admin")
	{
		a.POST("/register", admin.Register)
	}

	wechat := handle.NewWechat(cfg.WechatToken, appSvc, authorSvc, bookSvc)
	wx := r.Group("/wx")
	{
		wx.GET("/", wechat.CheckSignature)
		wx.POST("/", wechat.Dispatch)
	}

}
