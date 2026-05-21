package domain

import (
	airportDomain "api/internal/airport/domain"
	flightDomain "api/internal/flight/domain"
	"shared/common"
	"time"
)

type Payload struct {
	Email common.Email

	ScheduledDeparture time.Time
	ActualDeparture    *time.Time
	FlightStatus       flightDomain.FlightStatus

	DepartureAirportIATACode airportDomain.IATACode
	DepartureAirportTitle    airportDomain.Title
	DepartureAirportCity     common.City
	DepartureAirportCountry  common.Country

	ArrivalAirportIATACode airportDomain.IATACode
	ArrivalAirportTitle    airportDomain.Title
	ArrivalAirportCity     common.City
	ArrivalAirportCountry  common.Country
}

func NewPayload(
	email common.Email,
	flight flightDomain.Flight,
	depAirport,
	arrAirport airportDomain.Airport,
) (Payload, error) {
	return Payload{
		Email: email,

		ScheduledDeparture: flight.ScheduledDeparture,
		ActualDeparture:    flight.ActualDeparture,
		FlightStatus:       flight.Status,

		DepartureAirportIATACode: depAirport.IATACode,
		DepartureAirportTitle:    depAirport.Title,
		DepartureAirportCity:     depAirport.City,
		DepartureAirportCountry:  depAirport.Country,

		ArrivalAirportIATACode: arrAirport.IATACode,
		ArrivalAirportTitle:    arrAirport.Title,
		ArrivalAirportCity:     arrAirport.City,
		ArrivalAirportCountry:  arrAirport.Country,
	}, nil
}
