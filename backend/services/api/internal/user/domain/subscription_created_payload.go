package domain

import (
	airportDomain "api/internal/airport/domain"
	flightDomain "api/internal/flight/domain"
	"encoding/json"
	"shared/common"
	"time"
)

// INFO: SubscriptionCreatedPayload structure is topic specific
type SubscriptionCreatedPayload struct {
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

func NewSubscriptionCreatedPayload(
	email common.Email,
	flight flightDomain.Flight,
	depAirport,
	arrAirport airportDomain.Airport,
) (SubscriptionCreatedPayload, error) {
	return SubscriptionCreatedPayload{
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

func (p *SubscriptionCreatedPayload) MarshalJSON() ([]byte, error) {
	m := SubscriptionCreatedPayloadModel{
		Email:                    p.Email.String(),
		ScheduledDeparture:       p.ScheduledDeparture,
		ActualDeparture:          p.ActualDeparture,
		FlightStatus:             p.FlightStatus.String(),
		DepartureAirportIATACode: p.DepartureAirportIATACode.String(),
		DepartureAirportTitle:    p.DepartureAirportTitle.String(),
		DepartureAirportCity:     p.DepartureAirportCity.String(),
		DepartureAirportCountry:  p.DepartureAirportCountry.String(),
		ArrivalAirportIATACode:   p.ArrivalAirportIATACode.String(),
		ArrivalAirportTitle:      p.ArrivalAirportTitle.String(),
		ArrivalAirportCity:       p.ArrivalAirportCity.String(),
		ArrivalAirportCountry:    p.ArrivalAirportCountry.String(),
	}
	return json.Marshal(m)
}

func (p *SubscriptionCreatedPayload) UnmarshalJSON(data []byte) error {
	m := &SubscriptionCreatedPayloadModel{}
	if err := json.Unmarshal(data, m); err != nil {
		return err
	}

	email, err := common.NewEmail(m.Email)
	if err != nil {
		return err
	}
	fs, err := flightDomain.NewFlightStatus(m.FlightStatus)

	depIATA, err := airportDomain.NewIATACode(m.DepartureAirportIATACode)
	if err != nil {
		return err
	}
	depTitle, err := airportDomain.NewTitle(m.DepartureAirportTitle)
	if err != nil {
		return err
	}
	depCity, err := common.NewCity(m.DepartureAirportCity)
	if err != nil {
		return err
	}
	depCountry, err := common.NewCountry(m.DepartureAirportCountry)
	if err != nil {
		return err
	}

	arrIATA, err := airportDomain.NewIATACode(m.ArrivalAirportIATACode)
	if err != nil {
		return err
	}
	arrTitle, err := airportDomain.NewTitle(m.ArrivalAirportTitle)
	if err != nil {
		return err
	}
	arrCity, err := common.NewCity(m.ArrivalAirportCity)
	if err != nil {
		return err
	}
	arrCountry, err := common.NewCountry(m.ArrivalAirportCountry)
	if err != nil {
		return err
	}

	*p = SubscriptionCreatedPayload{
		Email: email,

		ScheduledDeparture: m.ScheduledDeparture,
		ActualDeparture:    m.ActualDeparture,

		FlightStatus: fs,

		DepartureAirportIATACode: depIATA,
		DepartureAirportTitle:    depTitle,
		DepartureAirportCity:     depCity,
		DepartureAirportCountry:  depCountry,

		ArrivalAirportIATACode: arrIATA,
		ArrivalAirportTitle:    arrTitle,
		ArrivalAirportCity:     arrCity,
		ArrivalAirportCountry:  arrCountry,
	}
	return nil
}
