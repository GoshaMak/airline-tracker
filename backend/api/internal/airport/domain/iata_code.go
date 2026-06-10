package domain

import (
	"fmt"
	"regexp"
)

type IATACode string

var iataCodeRegexp = regexp.MustCompile(`^[A-Z]{3}$`)

// TODO: better matching
func NewIATACode(v string) (IATACode, error) {
	if !iataCodeRegexp.MatchString(v) {
		return "", fmt.Errorf("invalid IATA code: %s", v)
	}
	return IATACode(v), nil
}

func (c IATACode) String() string {
	return string(c)
}
