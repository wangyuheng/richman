package main

import (
	"log"

	"github.com/geeklubcn/richman/config"
	"github.com/geeklubcn/richman/handle"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	//env()

	cfg := config.Load()
	logrus.SetLevel(cfg.LogLevel)

	logrus.Debugf("load config. %+v", cfg)
	r := gin.New()
	register(r)
	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}

func register(r *gin.Engine) {
	cfg := config.GetConfig()
	feishu := handle.NewFeishu(cfg.AppId, cfg.AppSecret)

	f := r.Group("/feishu")
	{
		f.POST("/webhook/:app_id", feishu.Webhook)
		f.POST("/register", feishu.Register)
	}

}
