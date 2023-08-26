package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/usecase"
)

type DevboxHandler interface {
	GetUserByID(ctx *gin.Context)
	PreparedLedger(ctx *gin.Context)
}

type devboxHandler struct {
	user   usecase.UserUseCase
	ledger usecase.LedgerUseCase
}

func NewDevboxHandler(cfg *config.Config, user usecase.UserUseCase, ledger usecase.LedgerUseCase) DevboxHandler {
	return &devboxHandler{user: user, ledger: ledger}
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

func (d *devboxHandler) PreparedLedger(ctx *gin.Context) {
	ctx.JSON(200, d.ledger.PreparedAllocated())
}
