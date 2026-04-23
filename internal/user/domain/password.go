package domain

import (
	"errors"
	"regexp"
)

type Password string

var (
	ErrShortPassword   = errors.New("password must contain at least 8 letters")
	ErrLongPassword    = errors.New("password is too long")
	ErrInvalidPassword = errors.New("invalid password")
)

const (
	minLen = 8
	maxLen = 100
)

var passwordRegex = regexp.MustCompile(`^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,100}$`)

func IsValidPassword(p string) error {
	if len(p) < minLen {
		return ErrShortPassword
	}
	if len(p) > maxLen {
		return ErrLongPassword
	}
	if !passwordRegex.MatchString(p) {
		return ErrInvalidPassword
	}
	return nil
}

func NewPassword(password string) (Password, error) {
	if err := IsValidPassword(password); err != nil {
		return "", err
	}
	return Password(password), nil
}
