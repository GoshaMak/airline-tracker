package command

import (
	"encoding/json"
	"notifier/internal/receiver/domain"
	"time"
)

type PayloadModel struct {
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
	c := PayloadModel{
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

	payload, err := json.Marshal(&c)
	if err != nil {
		return domain.Notification{}, err
	}

	return domain.NewNotification(payload, sendAt, status)
}
