package http

import (
	"github.com/gin-gonic/gin"
	"github.com/wangyuheng/richman/internal/interfaces/http/handler"
)

func NewEngine(wh handler.WechatHandler, dev handler.DevboxHandler) *gin.Engine {
	router := gin.Default()

	router.GET("", func(ctx *gin.Context) {
		ctx.Redirect(302, "https://github.com/wangyuheng/richman")
		return
	})

	devbox := router.Group("/devbox")
	{
		devbox.Any("GetUserByID", dev.GetUserByID)
		devbox.Any("PreparedLedger", dev.PreparedLedger)
	}

	v2 := router.Group("/v2")
	{
		wx := v2.Group("/wx")
		{
			wx.GET("", wh.CheckSignature)
			wx.POST("", wh.Dispatch)
		}
	}

	return router
}
