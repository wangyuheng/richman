package main

import (
	"github.com/geeklubcn/feishu-bitable-db/db"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/wangyuheng/richman/internal/infrastructure/database"
	"log"
	"net/http"
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

	bdb, err := db.NewDB(cfg.DbAppId, cfg.DbAppSecret)
	if err != nil {
		panic(err)
	}

	larkCli := lark.NewClient(cfg.DbAppId, cfg.DbAppSecret,
		lark.WithLogLevel(larkcore.LogLevelDebug),
		lark.WithReqTimeout(100*time.Second),
		lark.WithHttpClient(http.DefaultClient))

	auditLogger := database.NewAuditLogService(cfg, bdb)

	r, err := InitializeEngine(cfg, bdb, larkCli, auditLogger)
	if err != nil {
		panic(err)
	}
	pprof.Register(r)
	r.Use(requestid.New())

	logger, _ := zap.NewProduction()
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	if err = r.Run(); err != nil {
		log.Fatal(err)
	}
}
