package command

import (
	"encoding/json"
	"notifier/internal/receiver/domain"
	"time"

	"github.com/google/uuid"
)

type SubscriptionCreatedPayloadModel struct {
	Email string `json:"email"`

	ScheduledDeparture time.Time  `json:"scheduled_departure"`
	ActualDeparture    *time.Time `json:"actual_departure"`
	FlightStatus       string     `json:"flight_status"`

	DepartureAirportIATACode string `json:"departure_airport_iata_code"`
	DepartureAirportTitle    string `json:"departure_airport_title"`
	DepartureAirportCity     string `json:"departure_airport_city"`
	DepartureAirportCountry  string `json:"departure_airport_country"`

	ArrivalAirportIATACode string `json:"arrival_airport_iata_code"`
	ArrivalAirportTitle    string `json:"arrival_airport_title"`
	ArrivalAirportCity     string `json:"arrival_airport_city"`
	ArrivalAirportCountry  string `json:"arrival_airport_country"`
}

func SubscriptionCreatedCommandToDomain(
	cmd *SubscriptionCreatedCommand,
	sendAt time.Time,
	status domain.NotificationStatus,
) (domain.Notification, error) {
	pm := SubscriptionCreatedPayloadModel{
		Email: cmd.Email.String(),

		ScheduledDeparture: cmd.ScheduledDeparture,
		ActualDeparture:    cmd.ActualDeparture,

		FlightStatus: cmd.FlightStatus,

		DepartureAirportIATACode: cmd.DepartureAirportIATACode,
		DepartureAirportTitle:    cmd.DepartureAirportTitle,
		DepartureAirportCity:     cmd.DepartureAirportCity.String(),
		DepartureAirportCountry:  cmd.DepartureAirportCountry.String(),

		ArrivalAirportIATACode: cmd.ArrivalAirportIATACode,
		ArrivalAirportTitle:    cmd.ArrivalAirportTitle,
		ArrivalAirportCity:     cmd.ArrivalAirportCity.String(),
		ArrivalAirportCountry:  cmd.ArrivalAirportCountry.String(),
	}

	payload, err := json.Marshal(&pm)
	if err != nil {
		return domain.Notification{}, err
	}

	return domain.NewNotification(payload, sendAt, status, domain.NotificationSubscribed)
}

type FlightUpdatedPayloadModel struct {
	FlightId uuid.UUID `json:"flight_id"`
	Users    []string  `json:"users,omitempty"`

	DepartureAirportTitle string `json:"departure_airport_title"`
	ArrivalAirportTitle   string `json:"arrival_airport_title"`

	ScheduledDeparture *time.Time `json:"scheduled_departure,omitempty"`
	ActualDeparture    *time.Time `json:"actual_departure,omitempty"`

	ScheduledArrival *time.Time `json:"scheduled_arrival,omitempty"`
	ActualArrival    *time.Time `json:"actual_arrival,omitempty"`

	Status *string `json:"status,omitempty"`
	Plan   *string `json:"plan,omitempty"`
}

func FlightUpdatedCommandToDomain(
	cmd FlightUpdatedCommand,
	sendAt time.Time,
	status domain.NotificationStatus,
) (domain.Notification, error) {
	users := make([]string, len(cmd.Users))
	for i, u := range cmd.Users {
		users[i] = u.String()
	}
	pm := FlightUpdatedPayloadModel{
		FlightId:              cmd.FlightId,
		Users:                 users,
		DepartureAirportTitle: cmd.DepartureAirportTitle,
		ArrivalAirportTitle:   cmd.ArrivalAirportTitle,
		ScheduledDeparture:    cmd.ScheduledDeparture,
		ActualDeparture:       cmd.ActualDeparture,
		ScheduledArrival:      cmd.ScheduledArrival,
		ActualArrival:         cmd.ActualArrival,
		Status:                cmd.Status,
		Plan:                  cmd.Plan,
	}

	payload, err := json.Marshal(&pm)
	if err != nil {
		return domain.Notification{}, err
	}

	n, err := domain.NewNotification(payload, sendAt, status, domain.NotificationFlightUpdated)
	if err != nil {
		return domain.Notification{}, err
	}

	return n, nil
}
