package usecase

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/wangyuheng/richman/internal/domain"
)

type UserUseCase interface {
	GetByID(UID string) (*domain.User, bool)
	Save(user domain.User) error
}

type userUseCase struct {
	userRepository domain.UserRepository
}

func NewUserUseCase(userRepository domain.UserRepository) UserUseCase {
	return &userUseCase{
		userRepository: userRepository,
	}
}

func (u *userUseCase) GetByID(UID string) (*domain.User, bool) {
	return u.userRepository.GetByID(UID)
}

func (u *userUseCase) Save(it domain.User) error {
	ctx := context.Background()
	logger := logrus.WithContext(ctx).WithField("user", it)
	if _, exist := u.GetByID(it.UID); exist {
		return nil
	}
	_, err := u.userRepository.Save(&it)
	if err != nil {
		logger.WithError(err).Error("save user err")
	}

	return err
}
