package domain

import (
	"fmt"
	"regexp"
)

type IATACode string

func NewIATACode(v string) (IATACode, error) {
	matched, err := regexp.MatchString(`^[A-Z]{3}$`, v)
	if err != nil {
		return "", err
	}
	if !matched {
		return "", fmt.Errorf("invalid iata code")
	}
	return IATACode(v), nil
}
