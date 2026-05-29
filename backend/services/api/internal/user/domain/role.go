package domain

import (
	"errors"
)

var (
	ErrInvalidRole = errors.New("invalid role")
)

// Must be created via NewRole
type Role int

const (
	UserRole Role = iota
	AdminRole
)

const (
	user  = "user"
	admin = "admin"
)

func NewRole(v string) (Role, error) {
	switch v {
	case user:
		return UserRole, nil

	case admin:
		return AdminRole, nil

	default:
		return -1, ErrInvalidRole
	}
}

func (r Role) String() string {
	return [...]string{user, admin}[r]
}
