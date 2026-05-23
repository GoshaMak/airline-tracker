package command

import (
	"notifier/internal/receiver/dto"
	"time"
)

type SendNotificationCommand struct {
	To string

	Departure    time.Time
	FlightStatus string

	DepartureAirportIATACode string
	DepartureAirportTitle    string
	DepartureAirportCity     string
	DepartureAirportCountry  string

	ArrivalAirportIATACode string
	ArrivalAirportTitle    string
	ArrivalAirportCity     string
	ArrivalAirportCountry  string
}

func NewSendNotificationCommand(n dto.NotificationDTO) (SendNotificationCommand, error) {
	var dep time.Time
	if n.ActualDeparture != nil {
		dep = *n.ActualDeparture
	} else {
		dep = n.ScheduledDeparture
	}

	return SendNotificationCommand{
		To:                       n.Email,
		Departure:                dep,
		FlightStatus:             n.FlightStatus,
		DepartureAirportIATACode: n.DepartureAirportIATACode,
		DepartureAirportTitle:    n.DepartureAirportTitle,
		DepartureAirportCity:     n.DepartureAirportCity,
		DepartureAirportCountry:  n.DepartureAirportCountry,
		ArrivalAirportIATACode:   n.ArrivalAirportIATACode,
		ArrivalAirportTitle:      n.ArrivalAirportTitle,
		ArrivalAirportCity:       n.ArrivalAirportCity,
		ArrivalAirportCountry:    n.ArrivalAirportCountry,
	}, nil
}
