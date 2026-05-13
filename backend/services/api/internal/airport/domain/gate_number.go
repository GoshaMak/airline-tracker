package domain

import (
	"fmt"
	"regexp"
)

type GateNumber string

func NewGateNumber(v string) (GateNumber, error) {
	matched, err := regexp.MatchString(`^[A-Z]\d{1,4}$`, v)
	if err != nil {
		return "", err
	}
	if !matched {
		return "", fmt.Errorf("invalid gate number")
	}
	return GateNumber(v), nil
}

func (gn GateNumber) String() string {
	return string(gn)
}
