package command

import (
	"notifier/internal/receiver/dto"
	"shared/common"
	"time"
)

type SubscriptionCreatedCommand struct {
	Email common.Email

	ScheduledDeparture time.Time
	ActualDeparture    *time.Time
	FlightStatus       string

	DepartureAirportIATACode string
	DepartureAirportTitle    string
	DepartureAirportCity     common.City
	DepartureAirportCountry  common.Country

	ArrivalAirportIATACode string
	ArrivalAirportTitle    string
	ArrivalAirportCity     common.City
	ArrivalAirportCountry  common.Country
}

func NewSubscriptionCreatedCommand(
	req *dto.SubscriptionCreatedDTO,
) (SubscriptionCreatedCommand, error) {
	email, err := common.NewEmail(req.Email)
	if err != nil {
		return SubscriptionCreatedCommand{}, err
	}

	depCity, err := common.NewCity(req.DepartureAirportCity)
	if err != nil {
		return SubscriptionCreatedCommand{}, err
	}
	depCountry, err := common.NewCountry(req.DepartureAirportCountry)
	if err != nil {
		return SubscriptionCreatedCommand{}, err
	}

	arrCity, err := common.NewCity(req.ArrivalAirportCity)
	if err != nil {
		return SubscriptionCreatedCommand{}, err
	}
	arrCountry, err := common.NewCountry(req.ArrivalAirportCountry)
	if err != nil {
		return SubscriptionCreatedCommand{}, err
	}

	return SubscriptionCreatedCommand{
		Email: email,

		ScheduledDeparture: req.ScheduledDeparture,
		ActualDeparture:    req.ActualDeparture,

		FlightStatus: req.FlightStatus,

		DepartureAirportIATACode: req.DepartureAirportIATACode,
		DepartureAirportTitle:    req.DepartureAirportTitle,
		DepartureAirportCity:     depCity,
		DepartureAirportCountry:  depCountry,

		ArrivalAirportIATACode: req.ArrivalAirportIATACode,
		ArrivalAirportTitle:    req.ArrivalAirportTitle,
		ArrivalAirportCity:     arrCity,
		ArrivalAirportCountry:  arrCountry,
	}, nil
}
