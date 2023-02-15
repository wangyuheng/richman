package biz

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/model"
	"github.com/wangyuheng/richman/internal/repo"
)

type User interface {
	Unique(ctx context.Context, uid string) (*model.User, bool)
	Save(ctx context.Context, user model.User) error
}

type user struct {
	users repo.Users
}

func NewUser(_ *config.Config, users repo.Users) User {
	return &user{
		users: users,
	}
}

func (b user) Unique(_ context.Context, uid string) (*model.User, bool) {
	return b.users.QueryByUID(uid)
}

func (b user) Save(ctx context.Context, it model.User) error {
	logger := logrus.WithContext(ctx).WithField("user", it)

	_, err := b.users.Save(&it)
	if err != nil {
		logger.WithError(err).Error("save user err")
	}

	return err
}
