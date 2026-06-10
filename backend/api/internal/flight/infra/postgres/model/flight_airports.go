package model

import (
	"github.com/google/uuid"
)

type FlightAirportsModel struct {
	DepartureAirportID       uuid.UUID `db:"departure_airport_id"`
	DepartureAirportIATACode string    `db:"departure_airport_iata_code"`
	DepartureAirportTitle    string    `db:"departure_airport_title"`
	DepartureAirportCity     string    `db:"departure_airport_city"`
	DepartureAirportCountry  string    `db:"departure_airport_country"`

	ArrivalAirportID       uuid.UUID `db:"arrival_airport_id"`
	ArrivalAirportIATACode string    `db:"arrival_airport_iata_code"`
	ArrivalAirportTitle    string    `db:"arrival_airport_title"`
	ArrivalAirportCity     string    `db:"arrival_airport_city"`
	ArrivalAirportCountry  string    `db:"arrival_airport_country"`
}
