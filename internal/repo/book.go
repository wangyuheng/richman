package repo

import (
	"context"
	"fmt"
	"github.com/geeklubcn/feishu-bitable-db/db"
	"github.com/sirupsen/logrus"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/model"
	"sync"
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
	QueryByUID(uid string) (*model.Book, bool)
}

type books struct {
	db    db.DB
	cache sync.Map
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

func (b *books) WarmUP(ctx context.Context) {}

func (b *books) Key(s string) string {
	return fmt.Sprintf("repo:book:%s", s)
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

func (b *books) QueryByUID(uid string) (*model.Book, bool) {
	ctx := context.Background()

	if v, ok := b.cache.Load(b.Key(uid)); ok {
		if vv, ok := v.(*model.Book); ok {
			return vv, true
		}
	}

	rs := b.db.Read(ctx, bookDatabase, bookTable, []db.SearchCmd{
		{BookCreatorID, "=", uid},
	})
	if len(rs) == 0 {
		return nil, false
	}
	res := &model.Book{
		AppToken:    db.GetString(rs[0], BookAppToken),
		Name:        db.GetString(rs[0], BookName),
		URL:         db.GetString(rs[0], BookURL),
		CreatorID:   db.GetString(rs[0], BookCreatorID),
		CreatorName: db.GetString(rs[0], BookCreatorName),
	}
	b.cache.Store(b.Key(res.CreatorID), res)
	return res, true
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
