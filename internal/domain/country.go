package domain

import "fmt"

type Country string

func NewCountry(v string) (Country, error) {
	if v == "" {
		return "", fmt.Errorf("invalid country")
	}
	return Country(v), nil
}
