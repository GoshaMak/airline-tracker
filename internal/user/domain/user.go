package domain

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID
	Email    string
	Password string
	Role     string
}

func NewUser(email, password, role string) *User {
	return &User{
		ID:       uuid.New(),
		Email:    email,
		Password: password,
		Role:     role,
	}
}
