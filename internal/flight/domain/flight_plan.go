package domain

import (
	"errors"
	"regexp"
	"strings"
)

type FlightPlan string

var (
	ErrInvalidFlightPlan = errors.New("invalid flight plan")
)

var flightPlanRegex = regexp.MustCompile(`^[A-Z0-9\- ]{2,200}$`)

func IsValidFlightPlan(p string) error {
	p = strings.ToUpper(strings.TrimSpace(p))

	if p == "" {
		return nil
	}

	if len(p) > 1000 {
		return ErrInvalidFlightPlan
	}

	if !flightPlanRegex.MatchString(p) {
		return ErrInvalidFlightPlan
	}

	return nil
}

func NewFlightPlan(plan string) (FlightPlan, error) {
	if err := IsValidFlightPlan(plan); err != nil {
		return "", err
	}
	return FlightPlan(plan), nil
}

func (fp FlightPlan) String() string {
	return string(fp)
}
