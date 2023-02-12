package repo

import (
	"context"
	"github.com/geeklubcn/feishu-bitable-db/db"
	"github.com/sirupsen/logrus"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/model"
)

const (
	bookDatabase    = "Richman"
	bookTable       = "book"
	BookAppToken    = "app_token"
	BookName        = "name"
	BookURL         = "url"
	BookCreatorID   = "creator_id"
	BookCreatorName = "creator_name"
)

type Books interface {
	Save(it *model.Book) (string, error)
	Search(cmd []db.SearchCmd) []*model.Book
}

type books struct {
	db db.DB
}

func NewBooks(cfg *config.Config) Books {
	ctx := context.Background()
	it, err := db.NewDB(cfg.DbAppId, cfg.DbAppSecret)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("init repo err! db:%s, table:%s, cfg:%+v", bookDatabase, bookTable, cfg)
		return nil
	}
	_, _ = it.SaveTable(ctx, bookDatabase, db.Table{
		Name: bookTable,
		Fields: []db.Field{
			{Name: BookAppToken, Type: db.String},
			{Name: BookName, Type: db.String},
			{Name: BookURL, Type: db.String},
			{Name: BookCreatorID, Type: db.String},
			{Name: BookCreatorName, Type: db.String},
		},
	})
	return &books{db: it}
}

func (b *books) Save(it *model.Book) (string, error) {
	ctx := context.Background()

	for _, r := range b.db.Read(ctx, bookDatabase, bookTable, []db.SearchCmd{
		{BookAppToken, "=", it.AppToken},
	}) {
		_ = b.db.Delete(ctx, bookDatabase, bookTable, db.GetID(r))
	}

	return b.db.Create(ctx, bookDatabase, bookTable, map[string]interface{}{
		BookAppToken:    it.AppToken,
		BookName:        it.Name,
		BookURL:         it.URL,
		BookCreatorID:   it.CreatorID,
		BookCreatorName: it.CreatorName,
	})
}

func (b *books) Search(cmd []db.SearchCmd) []*model.Book {
	ctx := context.Background()
	rs := b.db.Read(ctx, bookDatabase, bookTable, cmd)
	res := make([]*model.Book, 0)
	for _, r := range rs {
		res = append(res, &model.Book{
			AppToken:    db.GetString(r, BookAppToken),
			Name:        db.GetString(r, BookName),
			URL:         db.GetString(r, BookURL),
			CreatorID:   db.GetString(r, BookCreatorID),
			CreatorName: db.GetString(r, BookCreatorName),
		})
	}
	return res
}
