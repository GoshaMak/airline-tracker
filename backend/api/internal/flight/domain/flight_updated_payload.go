package domain

import (
	airportDomain "api/internal/airport/domain"
	"encoding/json"
	"shared/common"
	"time"

	"github.com/google/uuid"
)

type FlightUpdatedPayload struct {
	FlightId uuid.UUID
	Users    []common.Email

	DepartureAirportTitle airportDomain.Title
	ArrivalAirportTitle   airportDomain.Title

	ScheduledDeparture *time.Time
	ActualDeparture    *time.Time

	ScheduledArrival *time.Time
	ActualArrival    *time.Time

	Status *string
	Plan   *string
}

func NewFlightUpdatedPayload(
	id uuid.UUID,
	users []common.Email,
	departureAirportTitle,
	arrivalAirportTitle airportDomain.Title,
	scheduledDeparture,
	actualDeparture,
	scheduledArrival,
	actualArrival *time.Time,
	status,
	plan *string,
) (FlightUpdatedPayload, error) {
	return FlightUpdatedPayload{
		FlightId:              id,
		Users:                 users,
		DepartureAirportTitle: departureAirportTitle,
		ArrivalAirportTitle:   arrivalAirportTitle,
		ScheduledDeparture:    scheduledDeparture,
		ActualDeparture:       actualDeparture,
		ScheduledArrival:      scheduledArrival,
		ActualArrival:         actualArrival,
		Status:                status,
		Plan:                  plan,
	}, nil
}

type FlightUpdatedPayloadModel struct {
	FlightId uuid.UUID `json:"flight_id"`
	Users    []string  `json:"users"`

	DepartureAirportTitle string `json:"departure_airport_title"`
	ArrivalAirportTitle   string `json:"arrival_airport_title"`

	ScheduledDeparture *time.Time `json:"scheduled_departure,omitempty"`
	ActualDeparture    *time.Time `json:"actual_departure,omitempty"`

	ScheduledArrival *time.Time `json:"scheduled_arrival,omitempty"`
	ActualArrival    *time.Time `json:"actual_arrival,omitempty"`

	Status *string `json:"status,omitempty"`
	Plan   *string `json:"plan,omitempty"`
}

func (f *FlightUpdatedPayload) MarshalJSON() ([]byte, error) {
	users := make([]string, len(f.Users))
	for i, u := range f.Users {
		users[i] = u.String()
	}
	m := FlightUpdatedPayloadModel{
		FlightId:              f.FlightId,
		DepartureAirportTitle: f.DepartureAirportTitle.String(),
		ArrivalAirportTitle:   f.ArrivalAirportTitle.String(),
		Users:                 users,
		ScheduledDeparture:    f.ScheduledDeparture,
		ActualDeparture:       f.ActualDeparture,
		ScheduledArrival:      f.ScheduledArrival,
		ActualArrival:         f.ActualArrival,
		Status:                f.Status,
		Plan:                  f.Plan,
	}
	return json.Marshal(&m)
}

func (f *FlightUpdatedPayload) UnmarshalJSON(data []byte) error {
	var m FlightUpdatedPayloadModel
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	emails := make([]common.Email, len(m.Users))
	for i, u := range m.Users {
		em, err := common.NewEmail(u)
		if err != nil {
			return err
		}
		emails[i] = em
	}
	depTitle, err := airportDomain.NewTitle(m.DepartureAirportTitle)
	if err != nil {
		return err
	}
	arrTitle, err := airportDomain.NewTitle(m.ArrivalAirportTitle)
	if err != nil {
		return err
	}
	pld, err := NewFlightUpdatedPayload(
		m.FlightId,
		emails,
		depTitle,
		arrTitle,
		m.ScheduledDeparture,
		m.ActualDeparture,
		m.ScheduledArrival,
		m.ActualArrival,
		m.Status,
		m.Plan,
	)
	if err != nil {
		return err
	}
	*f = pld
	return nil
}
