package api

import (
	"github.com/gin-gonic/gin"
)

type Router struct {
	Wechat Wechat
}

func (c Router) Register(r *gin.Engine) {
	v2 := r.Group("/v2")
	wx := v2.Group("/wx")
	{
		wx.GET("/", c.Wechat.CheckSignature)
		wx.POST("/", c.Wechat.Dispatch)
	}
}
