package command

import (
	"api/internal/flight/dto"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type FlightCommand struct {
	ScheduledDeparture time.Time
	ScheduledArrival   time.Time
	ActualDeparture    *time.Time
	ActualArrival      *time.Time
	Status             string
	Plan               *string
}

type CreateFlightCommand struct {
	Flight             FlightCommand
	AircraftId         uuid.UUID
	DepartureAiroprtId uuid.UUID
	ArrivalAiroprtId   uuid.UUID
	DepartureGateId    uuid.UUID
	ArrivalGateId      uuid.UUID
}

func NewCreateFlightCommand(req *dto.CreateFlightRequest) (CreateFlightCommand, error) {
	f, err := DTOToFlightCommand(req.Flight)
	if err != nil {
		return CreateFlightCommand{}, err
	}
	if f.ScheduledDeparture.Compare(f.ScheduledArrival) != -1 {
		return CreateFlightCommand{}, fmt.Errorf("arrival after departure")
	}
	if f.Status == "" {
		return CreateFlightCommand{}, fmt.Errorf("flight status is empty")
	}
	return CreateFlightCommand{
		Flight:             f,
		AircraftId:         req.AircraftId,
		DepartureAiroprtId: req.DepartureAirportId,
		ArrivalAiroprtId:   req.ArrivalAirportId,
		DepartureGateId:    req.DepartureGateId,
		ArrivalGateId:      req.ArrivalGateId,
	}, nil
}
