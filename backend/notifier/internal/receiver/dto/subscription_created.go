package dto

import "time"

type SubscriptionCreatedDTO struct {
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
