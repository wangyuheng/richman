package main

import (
	"log"
	"os"
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
	_ = os.Setenv("LARK_APP_ID", "cli_a218fd247d3b500c")
	_ = os.Setenv("LARK_APP_SECRET", "yOMK77lcLL1OmNmatJSGlgWwNG73KEiI")
	_ = os.Setenv("TEMPLATE_APP_TOKEN", "bascnVvTYC4C593vqcchbrDwMWc")
	_ = os.Setenv("TARGET_FOLDER_APP_TOKEN", "fldcntu3Hi1T0EBpBFkiMPM38Qb")
	_ = os.Setenv("AI_URL", "https://gpt.geeklub.cn/v1/chat/completions")
	_ = os.Setenv("AI_KEY", "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")
	_ = os.Setenv("WECHAT_TOKEN", "crick77")
	_ = os.Setenv("DB_APP_TOKEN", "bascnAxeGhi8uWchfxjlwsZdFId")
	_ = os.Setenv("DB_TABLE_TOKEN", "tblqW9kk0yc01dbp")

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
