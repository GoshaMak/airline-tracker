package common

import (
	"errors"
	"unicode"
)

// Must be created via NewCity
type City string

var (
	ErrInvalidCity = errors.New("invalid city")
)

func isValidCity(city string) bool {
	if len(city) < 2 || len(city) > 200 {
		return false
	}

	for _, r := range city {
		if !unicode.IsLetter(r) && r != ' ' && r != '-' {
			return false
		}
	}

	return true
}

func NewCity(c string) (City, error) {
	if !isValidCity(c) {
		return "", ErrInvalidCity
	}
	return City(c), nil
}
