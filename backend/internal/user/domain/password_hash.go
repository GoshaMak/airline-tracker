package domain

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

type PasswordHashed string

var (
	ErrShortPassword   = errors.New("password must contain at least 8 letters")
	ErrLongPassword    = errors.New("password is too long")
	ErrInvalidPassword = errors.New("invalid password")
)

const (
	minLen = 8
	maxLen = 100
)

var (
	lowercase  = regexp.MustCompile(`[a-z]`)
	uppercase  = regexp.MustCompile(`[A-Z]`)
	digit      = regexp.MustCompile(`\d`)
	special    = regexp.MustCompile(`[@$!%*?&]`)
	validChars = regexp.MustCompile(`^[A-Za-z\d@$!%*?&]{8,100}$`)
)

func IsValidPassword(p string) error {
	if len(p) < minLen {
		return ErrShortPassword
	}
	if len(p) > maxLen {
		return ErrLongPassword
	}
	if !validChars.MatchString(p) {
		return errors.New("invalid characters or length")
	}
	if !lowercase.MatchString(p) {
		return errors.New("must contain lowercase letter")
	}
	if !uppercase.MatchString(p) {
		return errors.New("must contain uppercase letter")
	}
	if !digit.MatchString(p) {
		return errors.New("must contain digit")
	}
	if !special.MatchString(p) {
		return errors.New("must contain special char")
	}
	return nil
}

func NewPasswordHashed(password string) (PasswordHashed, error) {
	if err := IsValidPassword(password); err != nil {
		return "", err
	}
	pswd, err := EncryptPassword(password)
	if err != nil {
		return "", err
	}

	return PasswordHashed(pswd), nil
}

func EncryptPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func (p PasswordHashed) String() string {
	return string(p)
}
