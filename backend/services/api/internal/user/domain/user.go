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

func NewUser(email, password, role string) (User, error) {
	e, err := common.NewEmail(email)
	if err != nil {
		return User{}, err
	}
	p, err := NewPasswordHashed(password)
	if err != nil {
		return User{}, err
	}
	r, err := NewRole(role)
	if err != nil {
		return User{}, err
	}
	return User{
		Id:           uuid.New(),
		Email:        e,
		PasswordHash: p,
		Role:         r,
	}, nil
}
