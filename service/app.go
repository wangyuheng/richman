package service

import (
	"context"

	"github.com/geeklubcn/feishu-bitable-db/db"
	"github.com/geeklubcn/richman/model"
)

const (
	appsDatabase          = "Richman"
	appsTable             = "apps"
	appsAppId             = "app_id"
	appsAppSecret         = "app_secret"
	appsVerificationToken = "verification_token"
)

type AppSvc interface {
	FindAll() []*model.App
	Save(m *model.App) (string, error)
}

func NewAppSvc(appId, appSecret string) AppSvc {
	ctx := context.Background()
	it, _ := db.NewDB(appId, appSecret)
	_, _ = it.SaveTable(ctx, appsDatabase, db.Table{
		Name: appsTable,
		Fields: []db.Field{
			{Name: appsAppId, Type: db.String},
			{Name: appsAppSecret, Type: db.String},
			{Name: appsVerificationToken, Type: db.String},
		},
	})
	return appSvc{it}
}

type appSvc struct {
	db db.DB
}

func (b appSvc) FindAll() []*model.App {
	ctx := context.Background()
	records := b.db.Read(ctx, appsDatabase, appsTable, nil)
	res := make([]*model.App, 0)
	for _, r := range records {
		res = append(res, &model.App{
			AppId:             db.GetString(r, appsAppId),
			AppSecret:         db.GetString(r, appsAppSecret),
			VerificationToken: db.GetString(r, appsVerificationToken),
		})
	}
	return res
}

func (b appSvc) Save(m *model.App) (string, error) {
	ctx := context.Background()

	for _, r := range b.db.Read(ctx, appsDatabase, appsTable, []db.SearchCmd{
		{appsAppId, "=", m.AppId},
	}) {
		_ = b.db.Delete(ctx, appsDatabase, appsTable, db.GetID(r))
	}

	return b.db.Create(ctx, appsDatabase, appsTable, map[string]interface{}{
		appsAppId:             m.AppId,
		appsAppSecret:         m.AppSecret,
		appsVerificationToken: m.VerificationToken,
	})
}
