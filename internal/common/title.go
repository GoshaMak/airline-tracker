package common

import (
	"errors"
	"unicode"
)

// Must be created via NewTitle
type Title string

var (
	ErrInvalidTitle = errors.New("invalid title")
)

func isValidTitle(t string) bool {
	if len(t) < 3 || len(t) > 200 {
		return false
	}

	for _, r := range t {
		if !unicode.IsLetter(r) && r != ' ' && r != '-' {
			return false
		}
	}

	return true
}

// TODO: parse from https://www.world-airport-codes.com/
func NewTitle(t string) (Title, error) {
	if !isValidTitle(t) {
		return "", ErrInvalidTitle
	}
	return Title(t), nil
}
