package command

import (
	"api/internal/flight/dto"
	"time"

	"github.com/google/uuid"
)

type UpdateFlightCommand struct {
	FlightId uuid.UUID

	ScheduledDeparture *time.Time
	ActualDeparture    *time.Time

	ScheduledArrival *time.Time
	ActualArrival    *time.Time

	Status *string
	Plan   *string
}

func NewUpdateFlightCommand(req *dto.UpdateFlightRequest) (UpdateFlightCommand, error) {
	return UpdateFlightCommand{
		FlightId:           req.Flight.FlightId,
		ScheduledDeparture: req.Flight.ScheduledDeparture,
		ActualDeparture:    req.Flight.ActualDeparture,
		ScheduledArrival:   req.Flight.ScheduledArrival,
		ActualArrival:      req.Flight.ActualArrival,
		Status:             req.Flight.Status,
		Plan:               req.Flight.Plan,
	}, nil
}
