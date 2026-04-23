package domain

import (
	"errors"
	"regexp"
	"strings"
)

type RegistrationNumber string

var (
	ErrInvalidRegistrationNumber = errors.New("invalid registration number")
)

var generalAircraftRegex = regexp.MustCompile(`^[A-Z0-9]{1,2}-[A-Z0-9]{2,5}$`)

func IsValidAircraftRegistration(reg string) bool {
	reg = strings.ToUpper(strings.TrimSpace(reg))
	return generalAircraftRegex.MatchString(reg)
}

func NewRegistrationNumber(n string) (RegistrationNumber, error) {
	if !IsValidAircraftRegistration(n) {
		return "", ErrInvalidRegistrationNumber
	}
	return RegistrationNumber(n), nil
}
