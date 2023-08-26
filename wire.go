//go:build wireinject
// +build wireinject

package main

import (
	"github.com/geeklubcn/feishu-bitable-db/db"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/domain"
	"github.com/wangyuheng/richman/internal/infrastructure/database"
	"github.com/wangyuheng/richman/internal/infrastructure/openai"
	"github.com/wangyuheng/richman/internal/interfaces/http"
	"github.com/wangyuheng/richman/internal/interfaces/http/handler"
	"github.com/wangyuheng/richman/internal/task"
	"github.com/wangyuheng/richman/internal/usecase"
)

func InitializeUserUseCase(db db.DB) (usecase.UserUseCase, error) {
	wire.Build(usecase.NewUserUseCase, database.NewUserRepository)
	return nil, nil
}

func InitializeLedgerUseCase(cfg *config.Config, db db.DB, larCli *lark.Client) (usecase.LedgerUseCase, error) {
	wire.Build(usecase.NewLedgerUseCase, database.NewLedgerRepository)
	return nil, nil
}

func InitializeBillUseCase(cfg *config.Config, db db.DB, larCli *lark.Client) (usecase.BillUseCase, error) {
	wire.Build(usecase.NewBillUseCase, database.NewBillRepository)
	return nil, nil
}

func InitializeWechatHandler(cfg *config.Config, db db.DB, larCli *lark.Client, auditLogger domain.AuditLogService) (handler.WechatHandler, error) {
	wire.Build(handler.NewWechatHandler, InitializeBillUseCase, InitializeUserUseCase, InitializeLedgerUseCase, openai.NewOpenAIService)
	return nil, nil
}

func InitializeDevboxHandler(cfg *config.Config, db db.DB, larCli *lark.Client) (handler.DevboxHandler, error) {
	wire.Build(handler.NewDevboxHandler, InitializeLedgerUseCase, InitializeUserUseCase)
	return nil, nil
}

func InitializeTask(cfg *config.Config, db db.DB, larCli *lark.Client) (task.Tasker, error) {
	wire.Build(task.NewWarmTask, InitializeLedgerUseCase, database.NewLedgerRepository, database.NewUserRepository)
	return nil, nil
}

func InitializeEngine(cfg *config.Config, db db.DB, larCli *lark.Client, auditLogger domain.AuditLogService) (*gin.Engine, error) {
	wire.Build(http.NewEngine, InitializeWechatHandler, InitializeDevboxHandler)
	return nil, nil
}

//var ComponentSet = wire.NewSet(
//	config.Load,
//	config.GetLarkConfig,
//	config.GetLarkDBConfig,
//	client.NewFeishu,
//	client.NewOpenAIService,
//)
//
//var ApiSet = wire.NewSet(
//	api.NewWechat,
//)
//
//var BizSet = wire.NewSet(
//	biz.NewBill,
//	biz.NewUser,
//	business.NewLedgerSvr,
//	business.NewFacade,
//)
//
//var RepoSet = wire.NewSet(
//	repo.NewBills,
//	repo.NewUsers,
//)
//
//func BuildRouter() api.Router {
//	panic(wire.Build(
//		wire.Struct(new(api.Router), "*"),
//		ComponentSet,
//		ApiSet,
//		BizSet,
//		RepoSet,
//	))
//}
