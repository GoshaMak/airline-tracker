package domain

import "fmt"

type IATACode string

func NewIATACode(v string) (IATACode, error) {
	if v == "" {
		return "", fmt.Errorf("invalid iata code")
	}
	return IATACode(v), nil
}
