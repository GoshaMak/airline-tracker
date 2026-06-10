package command

import (
	"notifier/internal/receiver/dto"
	"shared/common"
	"time"

	"github.com/google/uuid"
)

type FlightUpdatedCommand struct {
	FlightId uuid.UUID
	Users    []common.Email

	DepartureAirportTitle string
	ArrivalAirportTitle   string

	ScheduledDeparture *time.Time
	ActualDeparture    *time.Time

	ScheduledArrival *time.Time
	ActualArrival    *time.Time

	Status *string
	Plan   *string
}

func NewFlightUpdatedCommand(req *dto.FlightUpdatedDTO) (FlightUpdatedCommand, error) {
	emails := make([]common.Email, len(req.Users))
	for i, u := range req.Users {
		e, err := common.NewEmail(u)
		if err != nil {
			return FlightUpdatedCommand{}, err
		}
		emails[i] = e
	}
	return FlightUpdatedCommand{
		FlightId:              req.FlightId,
		Users:                 emails,
		DepartureAirportTitle: req.DepartureAirportTitle,
		ArrivalAirportTitle:   req.ArrivalAirportTitle,
		ScheduledDeparture:    req.ScheduledDeparture,
		ActualDeparture:       req.ActualDeparture,
		ScheduledArrival:      req.ScheduledArrival,
		ActualArrival:         req.ActualArrival,
		Status:                req.Status,
		Plan:                  req.Plan,
	}, nil
}
