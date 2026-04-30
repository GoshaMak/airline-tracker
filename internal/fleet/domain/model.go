package domain

import (
	"errors"
	"regexp"
	"strings"
)

type Model string

var (
	ErrInvalidModel = errors.New("invalid model")
)

var modelRegex = regexp.MustCompile(`^[A-Z0-9][A-Z0-9\s\-]{1,50}$`)

func IsValidModel(m string) bool {
	m = strings.ToUpper(strings.TrimSpace(m))
	return modelRegex.MatchString(m)
}

func NewModel(model string) (Model, error) {
	if !IsValidModel(model) {
		return "", ErrInvalidModel
	}
	return Model(model), nil
}

func (m Model) String() string {
	return string(m)
}
