package common

import (
	"errors"
	"net/mail"
	"strings"
)

var (
	ErrInvalidEmail = errors.New("invalid email")
)

// Must be created via NewEmail
type Email string

func NewEmail(v string) (Email, error) {
	v = strings.TrimSpace(v)
	addr, err := mail.ParseAddress(v)
	if err != nil {
		return "", ErrInvalidEmail
	}
	return Email(strings.ToLower(addr.Address)), nil
}

func (e Email) String() string {
	return string(e)
}
