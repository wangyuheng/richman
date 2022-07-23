package handle

import (
	"github.com/geeklubcn/richman/model"
	"github.com/geeklubcn/richman/service"
	"github.com/gin-gonic/gin"
)

type Admin interface {
	Register(ctx *gin.Context)
}

type admin struct {
	appSvc    service.AppSvc
	authorSvc service.AuthorSvc
	bookSvc   service.BookSvc
}

func NewAdmin(appSvc service.AppSvc, authorSvc service.AuthorSvc, bookSvc service.BookSvc) Admin {
	return &admin{appSvc, authorSvc, bookSvc}
}

func (a *admin) Register(ctx *gin.Context) {
	var app model.App
	err := ctx.ShouldBind(&app)
	if err != nil {
		_ = ctx.AbortWithError(500, err)
		return
	}
	_, err = a.appSvc.Save(&app)
	if err != nil {
		_ = ctx.AbortWithError(500, err)
		return
	}
	err = register(app, a.authorSvc, a.bookSvc)
	if err != nil {
		_ = ctx.AbortWithError(500, err)
		return
	}

	ctx.AbortWithStatus(200)
}
