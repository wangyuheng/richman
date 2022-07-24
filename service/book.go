package service

import (
	"context"
	"github.com/geeklubcn/richman/model"
	"github.com/sirupsen/logrus"

	"github.com/geeklubcn/feishu-bitable-db/db"
)

const (
	bookDatabase = "Richman"
	bookTable    = "books"
	bookAppId    = "app_id"
	bookAppToken = "app_token"
	bookOpenId   = "open_id"
	bookCategory = "category"
)

type BookSvc interface {
	GetByOpenId(openId string) (*model.Book, bool)
	Save(appId, openId, appToken, category string) (string, error)
}

type bookSvc struct {
	db db.DB
}

func NewBookSvc(appId, appSecret string) BookSvc {
	ctx := context.Background()
	it, err := db.NewDB(appId, appSecret)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("init DB err! appId:%s", appId)
		return nil
	}
	_, _ = it.SaveTable(ctx, bookDatabase, db.Table{
		Name: "books",
		Fields: []db.Field{
			{Name: bookAppId, Type: db.String},
			{Name: bookAppToken, Type: db.String},
			{Name: bookOpenId, Type: db.String},
			{Name: bookCategory, Type: db.String},
		},
	})
	return bookSvc{it}
}

func (b bookSvc) GetByOpenId(openId string) (*model.Book, bool) {
	ctx := context.Background()
	rs := b.db.Read(ctx, bookDatabase, bookTable, []db.SearchCmd{
		{bookOpenId, "=", openId},
	})
	for _, r := range rs {
		return &model.Book{
			AppId:    db.GetString(r, bookAppId),
			AppToken: db.GetString(r, bookAppToken),
			OpenId:   db.GetString(r, bookOpenId),
			Category: model.Category(db.GetString(r, bookCategory)),
		}, true
	}

	return nil, false
}

func (b bookSvc) Save(appId, openId, appToken, category string) (string, error) {
	ctx := context.Background()

	for _, r := range b.db.Read(ctx, bookDatabase, bookTable, []db.SearchCmd{
		{bookOpenId, "=", openId},
	}) {
		_ = b.db.Delete(ctx, bookDatabase, bookTable, db.GetID(r))
	}

	return b.db.Create(ctx, bookDatabase, bookTable, map[string]interface{}{
		bookAppId:    appId,
		bookAppToken: appToken,
		bookOpenId:   openId,
		bookCategory: category,
	})
}
