package api

import (
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

type Router struct {
	Wechat Wechat
}

func (r Router) Register(e *gin.Engine) {
	e.GET("", func(ctx *gin.Context) {
		ctx.Redirect(302, "https://github.com/wangyuheng/richman")
		return
	})
	// gpt proxy
	e.Any("/gpt", func(ctx *gin.Context) {
		gpt := "https://falling-base-15ce.wangyuheng.workers.dev/v1/chat/completions"
		target, _ := url.Parse(gpt)
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	})
	// richman
	v2 := e.Group("/v2")
	wx := v2.Group("/wx")
	{
		wx.GET("", r.Wechat.CheckSignature)
		wx.POST("", r.Wechat.Dispatch)
	}
}
