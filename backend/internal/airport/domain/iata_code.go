package domain

import (
	"errors"
	"regexp"
)

type IATACode string

var (
	ErrInvalidIATACode = errors.New("invalid IATA code")
)

// TODO: better matching
func NewIATACode(v string) (IATACode, error) {
	matched, err := regexp.MatchString(`^[A-Z]{3}$`, v)
	if err != nil || !matched {
		return "", ErrInvalidIATACode
	}
	return IATACode(v), nil
}

func (c IATACode) String() string {
	return string(c)
}
