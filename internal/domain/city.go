package domain

import "fmt"

type City string

func NewCity(v string) (City, error) {
	if v == "" {
		return "", fmt.Errorf("invalid country")
	}
	return City(v), nil
}
