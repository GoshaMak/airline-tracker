package domain

import (
	"errors"
	"unicode"
)

type Manufacturer string

var (
	ErrInvalidManufacturer = errors.New("invalid manufacturer")
)

func IsValidManufacturer(name string) bool {
	if len(name) < 2 || len(name) > 100 {
		return false
	}

	for _, r := range name {
		if !unicode.IsLetter(r) && r != ' ' && r != '-' {
			return false
		}
	}

	return true
}

func NewManufacturer(manufacturer string) (Manufacturer, error) {
	if !IsValidManufacturer(manufacturer) {
		return "", ErrInvalidManufacturer
	}
	return Manufacturer(manufacturer), nil
}
