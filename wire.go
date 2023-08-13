//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/api"
	"github.com/wangyuheng/richman/internal/biz"
	"github.com/wangyuheng/richman/internal/business"
	"github.com/wangyuheng/richman/internal/client"
	"github.com/wangyuheng/richman/internal/repo"
)

var ComponentSet = wire.NewSet(
	config.Load,
	config.GetLarkConfig,
	config.GetLarkDBConfig,
	client.NewFeishu,
	client.NewOpenAICaller,
)

var ApiSet = wire.NewSet(
	api.NewWechat,
)

var BizSet = wire.NewSet(
	biz.NewBill,
	biz.NewUser,
	business.NewLedgerSvr,
)

var RepoSet = wire.NewSet(
	repo.NewBills,
	repo.NewUsers,
)

func BuildRouter() api.Router {
	panic(wire.Build(
		wire.Struct(new(api.Router), "*"),
		ComponentSet,
		ApiSet,
		BizSet,
		RepoSet,
	))
}
