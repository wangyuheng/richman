package domain

type UserRepository interface {
	GetByID(UID string) (*User, bool)
	Save(user *User) (string, error)
}
