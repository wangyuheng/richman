// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/google/wire"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/api"
	"github.com/wangyuheng/richman/internal/biz"
	"github.com/wangyuheng/richman/internal/client"
	"github.com/wangyuheng/richman/internal/repo"
)

// Injectors from wire.go:

func BuildRouter() api.Router {
	configConfig := config.Load()
	bills := repo.NewBills(configConfig)
	bill := biz.NewBill(configConfig, bills)
	books := repo.NewBooks(configConfig)
	larkClient := client.NewFeishu(configConfig)
	book := biz.NewBook(configConfig, books, larkClient)
	users := repo.NewUsers(configConfig)
	user := biz.NewUser(configConfig, users)
	wechat := api.NewWechat(configConfig, bill, book, user)
	router := api.Router{
		Wechat: wechat,
	}
	return router
}

// wire.go:

var ComponentSet = wire.NewSet(config.Load, client.NewFeishu)

var ApiSet = wire.NewSet(api.NewWechat)

var BizSet = wire.NewSet(biz.NewBill, biz.NewBook, biz.NewUser)

var RepoSet = wire.NewSet(repo.NewBills, repo.NewBooks, repo.NewUsers)
