package command

import (
	"airline-tracker/internal/flight/dto"
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

type AddFlightCommand struct {
	Flight             FlightCommand
	AircraftID         uuid.UUID
	DepartureAiroprtID uuid.UUID
	ArrivalAiroprtID   uuid.UUID
	DepartureGateID    uuid.UUID
	ArrivalGateID      uuid.UUID
}

func NewAddFlightCommand(req *dto.CreateFlightRequest) (*AddFlightCommand, error) {
	f, err := DTOToFlightCommand(&req.Flight)
	if err != nil {
		return nil, err
	}
	if f.ScheduledDeparture.Compare(f.ScheduledArrival) != -1 {
		return nil, fmt.Errorf("arrival after departure")
	}
	if f.Status == "" {
		return nil, fmt.Errorf("flight status is empty")
	}
	return &AddFlightCommand{
		Flight:             f,
		AircraftID:         req.AircraftID,
		DepartureAiroprtID: req.DepartureAirportID,
		ArrivalAiroprtID:   req.ArrivalAirportID,
		DepartureGateID:    req.DepartureGateID,
		ArrivalGateID:      req.ArrivalGateID,
	}, nil
}
