package client

import (
	"github.com/larksuite/oapi-sdk-go/v3"
	"github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/wangyuheng/richman/config"
	"net/http"
	"time"
)

func NewFeishu(cfg *config.Config) *lark.Client {
	return lark.NewClient(cfg.DbAppId, cfg.DbAppSecret,
		lark.WithLogLevel(larkcore.LogLevelDebug),
		lark.WithReqTimeout(100*time.Second),
		lark.WithHttpClient(http.DefaultClient))
}
