package domain

import (
	"shared/common"

	"github.com/google/uuid"
)

type User struct {
	Id           uuid.UUID
	Email        common.Email
	PasswordHash PasswordHashed
	Role         Role
}

func NewUser(email, password string, role Role) (User, error) {
	e, err := common.NewEmail(email)
	if err != nil {
		return User{}, err
	}
	p, err := NewPasswordHashed(password)
	if err != nil {
		return User{}, err
	}
	return User{
		Id:           uuid.New(),
		Email:        e,
		PasswordHash: p,
		Role:         role,
	}, nil
}
