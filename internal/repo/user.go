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
	userDatabase = "bascnAxeGhi8uWchfxjlwsZdFId"
	userTable    = "tblmwBl2IVO3okT9"
	userUid      = "uid"
	userName     = "name"
)

type Users interface {
	Save(it *model.User) (string, error)
	QueryByUID(uid string) (*model.User, bool)
}

type users struct {
	db    db.DB
	cache sync.Map
}

func NewUsers(cfg *config.Config) Users {
	ctx := context.Background()
	it, err := db.NewDB(cfg.DbAppId, cfg.DbAppSecret)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("init repo err! db:%s, table:%s, cfg:%+v", userDatabase, userTable, cfg)
		return nil
	}
	_, _ = it.SaveTable(ctx, userDatabase, db.Table{
		Name: userTable,
		Fields: []db.Field{
			{Name: userUid, Type: db.String},
			{Name: userName, Type: db.String},
		},
	})
	return &users{db: it}
}

func (b *users) WarmUP(ctx context.Context) {}

func (b *users) Key(s string) string {
	return fmt.Sprintf("repo:user:%s", s)
}

func (b *users) Save(it *model.User) (string, error) {
	ctx := context.Background()

	for _, r := range b.db.Read(ctx, userDatabase, userTable, []db.SearchCmd{
		{userUid, "=", it.UID},
	}) {
		_ = b.db.Delete(ctx, userDatabase, userTable, db.GetID(r))
	}

	res, err := b.db.Create(ctx, userDatabase, userTable, map[string]interface{}{
		userUid:  it.UID,
		userName: it.Name,
	})
	if err == nil {
		b.cache.Store(b.Key(it.UID), res)
	}
	return res, err
}

func (b *users) QueryByUID(uid string) (*model.User, bool) {
	ctx := context.Background()

	if v, ok := b.cache.Load(b.Key(uid)); ok {
		if vv, ok := v.(*model.User); ok {
			return vv, true
		}
	}

	rs := b.db.Read(ctx, userDatabase, userTable, []db.SearchCmd{
		{userUid, "=", uid},
	})
	if len(rs) == 0 {
		return nil, false
	}
	res := &model.User{
		UID:  db.GetString(rs[0], userUid),
		Name: db.GetString(rs[0], userName),
	}
	b.cache.Store(b.Key(res.UID), res)
	return res, true
}
