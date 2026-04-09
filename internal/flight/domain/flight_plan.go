package domain

import "fmt"

type FlightPlan string

func NewFlightPlan(v string) (FlightPlan, error) {
	if v == "" {
		return "", fmt.Errorf("invalid flight plan")
	}
	return FlightPlan(v), nil
}

func (fp FlightPlan) String() string {
	return string(fp)
}
