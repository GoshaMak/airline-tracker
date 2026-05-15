package domain

import (
	"shared/common"
	"time"
)

type EmailMessage struct {
	To common.Email `json:"email"`

	Departure    time.Time `json:"departure"`
	FlightStatus string    `json:"flight_status"`

	DepartureAirportIATACode string         `json:"departure_airport_iata_code"`
	DepartureAirportTitle    string         `json:"departure_airport_title"`
	DepartureAirportCity     common.City    `json:"departure_airport_city"`
	DepartureAirportCountry  common.Country `json:"departure_airport_country"`

	ArrivalAirportIATACode string         `json:"arrival_airport_iata_code"`
	ArrivalAirportTitle    string         `json:"arrival_airport_title"`
	ArrivalAirportCity     common.City    `json:"arrival_airport_city"`
	ArrivalAirportCountry  common.Country `json:"arrival_airport_country"`
}

func NewEmailMessage(
	to string,
	departure time.Time,
	flightStatus,
	departureAirportIATACode,
	departureAirportTitle,
	departureAirportCity,
	departureAirportCountry,
	arrivalAirportIATACode,
	arrivalAirportTitle,
	arrivalAirportCity,
	arrivalAirportCountry string,
) (EmailMessage, error) {
	toEmail, err := common.NewEmail(to)
	if err != nil {
		return EmailMessage{}, err
	}

	depCity, err := common.NewCity(departureAirportCity)
	if err != nil {
		return EmailMessage{}, err
	}

	depCountry, err := common.NewCountry(departureAirportCountry)
	if err != nil {
		return EmailMessage{}, err
	}

	arrCity, err := common.NewCity(arrivalAirportCity)
	if err != nil {
		return EmailMessage{}, err
	}

	arrCountry, err := common.NewCountry(arrivalAirportCountry)
	if err != nil {
		return EmailMessage{}, err
	}

	return EmailMessage{
		To:                       toEmail,
		Departure:                departure,
		FlightStatus:             flightStatus,
		DepartureAirportIATACode: departureAirportIATACode,
		DepartureAirportTitle:    departureAirportTitle,
		DepartureAirportCity:     depCity,
		DepartureAirportCountry:  depCountry,
		ArrivalAirportIATACode:   arrivalAirportIATACode,
		ArrivalAirportTitle:      arrivalAirportTitle,
		ArrivalAirportCity:       arrCity,
		ArrivalAirportCountry:    arrCountry,
	}, nil
}
