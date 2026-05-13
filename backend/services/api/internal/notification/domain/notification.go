package domain

import (
	airportDomain "api/internal/airport/domain"
	"api/internal/common"
	flightDomain "api/internal/flight/domain"
	"strings"
	"time"
)

type Notification struct {
	Email common.Email

	ScheduledDeparture time.Time
	ActualDeparture    *time.Time
	FlightStatus       flightDomain.FlightStatus

	DepartureAirportIATACode string
	DepartureAirportTitle    string
	DepartureAirportCity     string
	DepartureAirportCountry  string

	ArrivalAirportIATACode string
	ArrivalAirportTitle    string
	ArrivalAirportCity     string
	ArrivalAirportCountry  string
}

func NewNotification(
	email common.Email,
	flight flightDomain.Flight,
	depAirport, arrAirport airportDomain.Airport,
) (Notification, error) {
	return Notification{
		Email: email,

		ScheduledDeparture: flight.ScheduledDeparture,
		ActualDeparture:    flight.ActualDeparture,
		FlightStatus:       flight.Status,

		DepartureAirportIATACode: depAirport.IATACode.String(),
		DepartureAirportTitle:    depAirport.Title.String(),
		DepartureAirportCity:     depAirport.City.String(),
		DepartureAirportCountry:  depAirport.City.String(),

		ArrivalAirportIATACode: arrAirport.IATACode.String(),
		ArrivalAirportTitle:    arrAirport.Title.String(),
		ArrivalAirportCity:     arrAirport.City.String(),
		ArrivalAirportCountry:  arrAirport.City.String(),
	}, nil
}

func (n Notification) String() string {
	var actDep string
	if n.ActualDeparture != nil {
		actDep = n.ActualDeparture.String()
	}
	return strings.Join(
		[]string{
			n.Email.String(),

			n.ScheduledDeparture.String(),
			actDep,
			n.FlightStatus.String(),

			n.DepartureAirportIATACode,
			n.DepartureAirportTitle,
			n.DepartureAirportCity,
			n.DepartureAirportCountry,

			n.ArrivalAirportIATACode,
			n.ArrivalAirportTitle,
			n.ArrivalAirportCity,
			n.ArrivalAirportCountry,
		},
		"\n",
	)
}
