package main

import (
	"log"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"go.uber.org/zap"

	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/wangyuheng/richman/config"
)

func main() {
	cfg := config.Load()
	logrus.SetLevel(cfg.LogLevel)

	logrus.Debugf("load config. %+v", cfg)
	r := gin.New()
	pprof.Register(r)
	r.Use(requestid.New())

	logger, _ := zap.NewProduction()
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	BuildRouter().Register(r)
	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}
