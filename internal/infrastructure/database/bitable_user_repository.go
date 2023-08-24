package database

import (
	"context"
	"fmt"
	"github.com/geeklubcn/feishu-bitable-db/db"
	"github.com/wangyuheng/richman/internal/domain"
	"sync"
)

const (
	userDatabase = "bascnAxeGhi8uWchfxjlwsZdFId"
	userTable    = "tblmwBl2IVO3okT9"
	userUid      = "uid"
	userName     = "name"
)

type userRepository struct {
	db    db.DB
	cache sync.Map
}

func NewUserRepository(db db.DB) domain.UserRepository {
	u := &userRepository{
		db: db,
	}
	u.WarmUP(context.Background())
	return u
}

func (u *userRepository) WarmUP(ctx context.Context) {
	items := u.db.Read(ctx, userDatabase, userTable, []db.SearchCmd{})
	for _, it := range items {
		res := &domain.User{
			UID:  db.GetString(it, userUid),
			Name: db.GetString(it, userName),
		}
		u.cache.Store(u.Key(res.UID), res)
	}
}

func (u *userRepository) GetByID(UID string) (*domain.User, bool) {
	ctx := context.Background()

	if v, ok := u.cache.Load(u.Key(UID)); ok {
		if vv, ok := v.(*domain.User); ok {
			return vv, true
		}
	}

	rs := u.db.Read(ctx, userDatabase, userTable, []db.SearchCmd{
		{userUid, "=", UID},
	})
	if len(rs) == 0 {
		return nil, false
	}
	res := &domain.User{
		UID:  db.GetString(rs[0], userUid),
		Name: db.GetString(rs[0], userName),
	}
	u.cache.Store(u.Key(res.UID), res)
	return res, true
}

func (u *userRepository) Save(it *domain.User) (string, error) {
	ctx := context.Background()

	for _, r := range u.db.Read(ctx, userDatabase, userTable, []db.SearchCmd{
		{userUid, "=", it.UID},
	}) {
		_ = u.db.Delete(ctx, userDatabase, userTable, db.GetID(r))
	}

	res, err := u.db.Create(ctx, userDatabase, userTable, map[string]interface{}{
		userUid:  it.UID,
		userName: it.Name,
	})
	if err == nil {
		u.cache.Store(u.Key(it.UID), res)
	}
	return res, err
}

func (u *userRepository) Key(s string) string {
	return fmt.Sprintf("cache:user:%s", s)
}
