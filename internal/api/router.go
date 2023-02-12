package api

import (
	"github.com/gin-gonic/gin"
)

type Router struct {
	Wechat Wechat
}

func (c Router) Register(r *gin.Engine) {
	wx := r.Group("/wx")
	{
		wx.GET("/", c.Wechat.CheckSignature)
		wx.POST("/", c.Wechat.Dispatch)
	}
}
