package domain

import "fmt"

type Title string

func NewTitle(v string) (Title, error) {
	if v == "" {
		return "", fmt.Errorf("invalid title")
	}
	return Title(v), nil
}
