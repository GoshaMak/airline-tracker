package domain

import (
	"errors"
)

type FlightStatus int

var (
	ErrInvalidStatus = errors.New("invalid flight status")
)

const (
	Scheduled FlightStatus = iota
	Boarding
	Departed
	Landed
	Arrived
	Delayed
	Cancelled
	Rescheduled
)

const (
	scheduled   = "scheduled"
	boarding    = "boarding"
	departed    = "departed"
	landed      = "landed"
	arrived     = "arrived"
	delayed     = "delayed"
	cancelled   = "cancelled"
	rescheduled = "rescheduled"
)

func NewFlightStatus(s string) (FlightStatus, error) {
	switch s {
	case scheduled:
		return Scheduled, nil
	case boarding:
		return Boarding, nil
	case departed:
		return Departed, nil
	case landed:
		return Landed, nil
	case arrived:
		return Arrived, nil
	case delayed:
		return Delayed, nil
	case cancelled:
		return Cancelled, nil
	case rescheduled:
		return Rescheduled, nil

	default:
		return -1, ErrInvalidStatus
	}
}

func (s FlightStatus) String() string {
	return [...]string{
		scheduled,
		boarding,
		departed,
		landed,
		arrived,
		delayed,
		cancelled,
		rescheduled,
	}[s]
}
