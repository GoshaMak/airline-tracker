package domain

import (
	"fmt"
	"regexp"
)

type GateNumber string

var gateNumberRegexp = regexp.MustCompile(`^[A-Z]\d{1,3}$`)

func NewGateNumber(v string) (GateNumber, error) {
	if !gateNumberRegexp.MatchString(v) {
		return "", fmt.Errorf("invalid gate number: %s", v)
	}
	return GateNumber(v), nil
}

func (gn GateNumber) String() string {
	return string(gn)
}
