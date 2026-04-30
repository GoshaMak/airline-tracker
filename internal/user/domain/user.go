package domain

import (
	"airline-tracker/internal/common"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        common.Email
	PasswordHash PasswordHashed
	Role         common.Role
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
	r, err := common.NewRole(role)
	if err != nil {
		return User{}, err
	}
	return User{
		ID:           uuid.New(),
		Email:        e,
		PasswordHash: p,
		Role:         r,
	}, nil
}
