package domain

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID
	Email    string
	Phone    string
	Password string
	Role     string
}

func NewUser(email, phone, password, role string) *User {
	return &User{
		ID:       uuid.New(),
		Email:    email,
		Phone:    phone,
		Password: password,
		Role:     role,
	}
}
