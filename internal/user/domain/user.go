package domain

import (
	"airline-tracker/internal/common"

	"github.com/google/uuid"
)

// WARN: fix this tag mess
type User struct {
	ID       uuid.UUID    `db:"id"`
	Email    common.Email `db:"email"`
	Password string       `db:"password"`
	Role     common.Role  `db:"role"`
}

func NewUser(email, password, role string) (User, error) {
	e, err := common.NewEmail(email)
	if err != nil {
		return User{}, err
	}
	r, err := common.NewRole(role)
	if err != nil {
		return User{}, err
	}
	return User{
		ID:       uuid.New(),
		Email:    e,
		Password: password,
		Role:     r,
	}, nil
}
