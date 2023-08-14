package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/usecase"
)

type DevboxHandler interface {
	GetUserByID(ctx *gin.Context)
}

type devboxHandler struct {
	user usecase.UserUseCase
}

func NewDevboxHandler(cfg *config.Config, user usecase.UserUseCase) DevboxHandler {
	return &devboxHandler{user: user}
}

func (d *devboxHandler) GetUserByID(ctx *gin.Context) {
	UID := ctx.Query("uid")
	user, exist := d.user.GetByID(UID)
	if exist {
		ctx.JSON(200, user)
	} else {
		ctx.JSON(400, fmt.Sprintf("User [%s] Not Found", UID))
	}
}
