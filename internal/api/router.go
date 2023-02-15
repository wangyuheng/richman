package api

import (
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
	v2 := e.Group("/v2")
	wx := v2.Group("/wx")
	{
		wx.GET("", r.Wechat.CheckSignature)
		wx.POST("", r.Wechat.Dispatch)
	}
}
