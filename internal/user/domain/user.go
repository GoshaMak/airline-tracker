package domain

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID `db:"id"`
	Email    string    `db:"email"`
	Password string    `db:"password"`
	Role     string    `db:"role"`
}

func NewUser(email, password, role string) *User {
	return &User{
		ID:       uuid.New(),
		Email:    email,
		Password: password,
		Role:     role,
	}
}
