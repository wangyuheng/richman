package repo

import (
	"context"

	"github.com/geeklubcn/richman/model"
	"github.com/sirupsen/logrus"

	"github.com/geeklubcn/feishu-bitable-db/db"
)

const (
	bookDatabase = "Richman"
	bookTable    = "books"
	BookAppId    = "app_id"
	BookAppToken = "app_token"
	BookOpenId   = "open_id"
	BookCategory = "category"
)

type Books interface {
	Search(cmd []db.SearchCmd) []*model.Book
	Save(it *model.Book) (string, error)
}

type books struct {
	db db.DB
}

func NewBooks(appId, appSecret string) Books {
	ctx := context.Background()
	it, err := db.NewDB(appId, appSecret)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("init DB err! appId:%s", appId)
		return nil
	}
	_, _ = it.SaveTable(ctx, bookDatabase, db.Table{
		Name: "books",
		Fields: []db.Field{
			{Name: BookAppId, Type: db.String},
			{Name: BookAppToken, Type: db.String},
			{Name: BookOpenId, Type: db.String},
			{Name: BookCategory, Type: db.String},
		},
	})
	return &books{db: it}
}

func (b *books) Search(cmd []db.SearchCmd) []*model.Book {
	ctx := context.Background()
	rs := b.db.Read(ctx, bookDatabase, bookTable, cmd)
	res := make([]*model.Book, 0)
	for _, r := range rs {
		res = append(res, &model.Book{
			AppId:    db.GetString(r, BookAppId),
			AppToken: db.GetString(r, BookAppToken),
			OpenId:   db.GetString(r, BookOpenId),
			Category: model.Category(db.GetString(r, BookCategory)),
		})
	}
	return res
}

func (b *books) Save(it *model.Book) (string, error) {
	ctx := context.Background()

	for _, r := range b.db.Read(ctx, bookDatabase, bookTable, []db.SearchCmd{
		{BookOpenId, "=", it.OpenId},
	}) {
		_ = b.db.Delete(ctx, bookDatabase, bookTable, db.GetID(r))
	}

	return b.db.Create(ctx, bookDatabase, bookTable, map[string]interface{}{
		BookAppId:    it.AppId,
		BookAppToken: it.AppToken,
		BookOpenId:   it.OpenId,
		BookCategory: it.Category,
	})
}
