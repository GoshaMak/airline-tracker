package domain

import "fmt"

type FlightStatus string

func NewFlightStatus(v string) (FlightStatus, error) {
	if v == "waiting" || v == "flying" || v == "finished" {
		return FlightStatus(v), nil
	}
	return "", fmt.Errorf("invalid flight status")

}

func (fs FlightStatus) String() string {
	return string(fs)
}
