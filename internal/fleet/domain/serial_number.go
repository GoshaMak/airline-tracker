package domain

import (
	"errors"
	"regexp"
	"strings"
)

type SerialNumber string

var (
	ErrInvalidSerialNumber = errors.New("invalid serial number")
)

var serialNumberRegex = regexp.MustCompile(`^[A-Z0-9][A-Z0-9\-\/]{1,49}$`)

func IsValidAircraftSerialNumber(serial string) bool {
	serial = strings.ToUpper(strings.TrimSpace(serial))
	return serialNumberRegex.MatchString(serial)
}

func NewSerialNumber(n string) (SerialNumber, error) {
	if !IsValidAircraftSerialNumber(n) {
		return "", ErrInvalidSerialNumber
	}
	return SerialNumber(n), nil
}

func (sn SerialNumber) String() string {
	return string(sn)
}
