package domain

import "context"

type UserRepository interface {
	GetByID(UID string) (*User, bool)
	Save(user *User) (string, error)
	WarmUP(ctx context.Context)
}
