package model

import (
	"api/internal/flight/domain"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type FlightDoc struct {
	ID                 string     `bson:"_id"`
	AircraftId         string     `bson:"aircraft_id"`
	ScheduledDeparture time.Time  `bson:"scheduled_departure"`
	ScheduledArrival   time.Time  `bson:"scheduled_arrival"`
	ActualDeparture    *time.Time `bson:"actual_departure,omitempty"`
	ActualArrival      *time.Time `bson:"actual_arrival,omitempty"`
	Status             string     `bson:"status"`
	Plan               *string    `bson:"plan,omitempty"`
	DepartureAirportId string     `bson:"departure_airport_id"`
	ArrivalAirportId   string     `bson:"arrival_airport_id"`
	DepartureGateId    string     `bson:"departure_gate_id"`
	ArrivalGateId      string     `bson:"arrival_gate_id"`
}

func ToFlightDoc(f domain.Flight) FlightDoc {
	var plan *string
	if f.Plan != nil {
		p := f.Plan.String()
		plan = &p
	}

	return FlightDoc{
		ID:                 f.Id.String(),
		AircraftId:         f.AircraftId.String(),
		ScheduledDeparture: f.ScheduledDeparture,
		ScheduledArrival:   f.ScheduledArrival,
		ActualDeparture:    f.ActualDeparture,
		ActualArrival:      f.ActualArrival,
		Status:             f.Status.String(),
		Plan:               plan,
		DepartureAirportId: f.DepartureAirportId.String(),
		ArrivalAirportId:   f.ArrivalAirportId.String(),
		DepartureGateId:    f.DepartureGateId.String(),
		ArrivalGateId:      f.ArrivalGateId.String(),
	}
}

func FromFlightDoc(doc FlightDoc) (domain.Flight, error) {
	slog.Debug("FromFlightDoc", "uuid", doc.ID)
	id, err := uuid.Parse(doc.ID)
	if err != nil {
		return domain.Flight{}, err
	}
	aircraftId, err := uuid.Parse(doc.AircraftId)
	if err != nil {
		return domain.Flight{}, err
	}
	depAirportId, err := uuid.Parse(doc.DepartureAirportId)
	if err != nil {
		return domain.Flight{}, err
	}
	arrAirportId, err := uuid.Parse(doc.ArrivalAirportId)
	if err != nil {
		return domain.Flight{}, err
	}
	depGateId, err := uuid.Parse(doc.DepartureGateId)
	if err != nil {
		return domain.Flight{}, err
	}
	arrGateId, err := uuid.Parse(doc.ArrivalGateId)
	if err != nil {
		return domain.Flight{}, err
	}

	status, err := domain.NewFlightStatus(doc.Status)
	if err != nil {
		return domain.Flight{}, err
	}

	var plan *domain.FlightPlan
	if doc.Plan != nil {
		p, err := domain.NewFlightPlan(*doc.Plan)
		if err != nil {
			return domain.Flight{}, err
		}
		plan = &p
	}

	return domain.Flight{
		Id:                 id,
		AircraftId:         aircraftId,
		ScheduledDeparture: doc.ScheduledDeparture,
		ScheduledArrival:   doc.ScheduledArrival,
		ActualDeparture:    doc.ActualDeparture,
		ActualArrival:      doc.ActualArrival,
		Status:             status,
		Plan:               plan,
		DepartureAirportId: depAirportId,
		ArrivalAirportId:   arrAirportId,
		DepartureGateId:    depGateId,
		ArrivalGateId:      arrGateId,
	}, nil
}

type SubscriptionDoc struct {
	FlightId string `bson:"flight_id"`
	UserId   string `bson:"user_id"`
}

type UserDoc struct {
	ID           string `bson:"_id"`
	Email        string `bson:"email"`
	PasswordHash string `bson:"password_hash"`
	Role         string `bson:"role"`
}
