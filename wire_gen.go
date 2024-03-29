// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/geeklubcn/feishu-bitable-db/db"
	"github.com/gin-gonic/gin"
	"github.com/larksuite/oapi-sdk-go/v3"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/domain"
	"github.com/wangyuheng/richman/internal/infrastructure/database"
	"github.com/wangyuheng/richman/internal/infrastructure/openai"
	"github.com/wangyuheng/richman/internal/interfaces/http"
	"github.com/wangyuheng/richman/internal/interfaces/http/handler"
	"github.com/wangyuheng/richman/internal/task"
	"github.com/wangyuheng/richman/internal/usecase"
)

// Injectors from wire.go:

func InitializeUserUseCase(db2 db.DB) (usecase.UserUseCase, error) {
	userRepository := database.NewUserRepository(db2)
	userUseCase := usecase.NewUserUseCase(userRepository)
	return userUseCase, nil
}

func InitializeLedgerUseCase(cfg *config.Config, db2 db.DB, larCli *lark.Client) (usecase.LedgerUseCase, error) {
	ledgerRepository := database.NewLedgerRepository(cfg, larCli, db2)
	ledgerUseCase := usecase.NewLedgerUseCase(cfg, ledgerRepository, larCli)
	return ledgerUseCase, nil
}

func InitializeBillUseCase(cfg *config.Config, db2 db.DB, larCli *lark.Client) (usecase.BillUseCase, error) {
	billRepository := database.NewBillRepository(db2)
	billUseCase := usecase.NewBillUseCase(billRepository)
	return billUseCase, nil
}

func InitializeWechatHandler(cfg *config.Config, db2 db.DB, larCli *lark.Client, auditLogger domain.AuditLogService) (handler.WechatHandler, error) {
	billUseCase, err := InitializeBillUseCase(cfg, db2, larCli)
	if err != nil {
		return nil, err
	}
	aiService := openai.NewOpenAIService(cfg, auditLogger)
	ledgerUseCase, err := InitializeLedgerUseCase(cfg, db2, larCli)
	if err != nil {
		return nil, err
	}
	userUseCase, err := InitializeUserUseCase(db2)
	if err != nil {
		return nil, err
	}
	wechatHandler := handler.NewWechatHandler(cfg, billUseCase, aiService, ledgerUseCase, userUseCase)
	return wechatHandler, nil
}

func InitializeDevboxHandler(cfg *config.Config, db2 db.DB, larCli *lark.Client) (handler.DevboxHandler, error) {
	userUseCase, err := InitializeUserUseCase(db2)
	if err != nil {
		return nil, err
	}
	ledgerUseCase, err := InitializeLedgerUseCase(cfg, db2, larCli)
	if err != nil {
		return nil, err
	}
	devboxHandler := handler.NewDevboxHandler(cfg, userUseCase, ledgerUseCase)
	return devboxHandler, nil
}

func InitializeTask(cfg *config.Config, db2 db.DB, larCli *lark.Client) (task.Tasker, error) {
	ledgerUseCase, err := InitializeLedgerUseCase(cfg, db2, larCli)
	if err != nil {
		return nil, err
	}
	userRepository := database.NewUserRepository(db2)
	ledgerRepository := database.NewLedgerRepository(cfg, larCli, db2)
	tasker := task.NewWarmTask(ledgerUseCase, userRepository, ledgerRepository)
	return tasker, nil
}

func InitializeEngine(cfg *config.Config, db2 db.DB, larCli *lark.Client, auditLogger domain.AuditLogService) (*gin.Engine, error) {
	wechatHandler, err := InitializeWechatHandler(cfg, db2, larCli, auditLogger)
	if err != nil {
		return nil, err
	}
	devboxHandler, err := InitializeDevboxHandler(cfg, db2, larCli)
	if err != nil {
		return nil, err
	}
	engine := http.NewEngine(wechatHandler, devboxHandler)
	return engine, nil
}
