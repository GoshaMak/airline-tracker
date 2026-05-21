package model

import (
	airportDomain "api/internal/airport/domain"
	flightDomain "api/internal/flight/domain"
	"api/internal/publisher/domain"
	"shared/common"
)

func OutboxDomainToModel(ob domain.Outbox) (OutboxModel, error) {
	payload, err := PayloadDomainToModel(ob.Payload)
	if err != nil {
		return OutboxModel{}, err
	}

	return OutboxModel{
		Id:        ob.Id,
		Topic:     ob.Topic,
		Payload:   payload,
		CreatedAt: ob.CreatedAt,
		SentAt:    ob.SentAt,
	}, nil
}

func PayloadDomainToModel(p domain.Payload) (PayloadModel, error) {
	return PayloadModel{
		Email: p.Email.String(),

		ScheduledDeparture: p.ScheduledDeparture,
		ActualDeparture:    p.ActualDeparture,

		FlightStatus: p.FlightStatus.String(),

		DepartureAirportIATACode: p.DepartureAirportIATACode.String(),
		DepartureAirportTitle:    p.DepartureAirportTitle.String(),
		DepartureAirportCity:     p.DepartureAirportCity.String(),
		DepartureAirportCountry:  p.DepartureAirportCountry.String(),

		ArrivalAirportIATACode: p.ArrivalAirportIATACode.String(),
		ArrivalAirportTitle:    p.ArrivalAirportTitle.String(),
		ArrivalAirportCity:     p.ArrivalAirportCity.String(),
		ArrivalAirportCountry:  p.ArrivalAirportCountry.String(),
	}, nil
}

func OutboxModelToDomain(obm OutboxModel) (domain.Outbox, error) {
	payload, err := PayloadModelToDomain(obm.Payload)
	if err != nil {
		return domain.Outbox{}, err
	}

	return domain.Outbox{
		Id:        obm.Id,
		Topic:     obm.Topic,
		Payload:   payload,
		CreatedAt: obm.CreatedAt,
		SentAt:    obm.SentAt,
	}, nil
}

func PayloadModelToDomain(pm PayloadModel) (domain.Payload, error) {
	email, err := common.NewEmail(pm.Email)
	if err != nil {
		return domain.Payload{}, err
	}

	fs, err := flightDomain.NewFlightStatus(pm.FlightStatus)
	if err != nil {
		return domain.Payload{}, err
	}

	depIATA, err := airportDomain.NewIATACode(pm.DepartureAirportIATACode)
	if err != nil {
		return domain.Payload{}, err
	}
	depTitle, err := airportDomain.NewTitle(pm.DepartureAirportTitle)
	if err != nil {
		return domain.Payload{}, err
	}
	depCity, err := common.NewCity(pm.DepartureAirportCity)
	if err != nil {
		return domain.Payload{}, err
	}
	depCountry, err := common.NewCountry(pm.DepartureAirportCountry)
	if err != nil {
		return domain.Payload{}, err
	}

	arrIATA, err := airportDomain.NewIATACode(pm.ArrivalAirportIATACode)
	if err != nil {
		return domain.Payload{}, err
	}
	arrTitle, err := airportDomain.NewTitle(pm.ArrivalAirportTitle)
	if err != nil {
		return domain.Payload{}, err
	}
	arrCity, err := common.NewCity(pm.ArrivalAirportCity)
	if err != nil {
		return domain.Payload{}, err
	}
	arrCountry, err := common.NewCountry(pm.ArrivalAirportCountry)
	if err != nil {
		return domain.Payload{}, err
	}

	return domain.Payload{
		Email: email,

		ScheduledDeparture: pm.ScheduledDeparture,
		ActualDeparture:    pm.ActualDeparture,

		FlightStatus: fs,

		DepartureAirportIATACode: depIATA,
		DepartureAirportTitle:    depTitle,
		DepartureAirportCity:     depCity,
		DepartureAirportCountry:  depCountry,

		ArrivalAirportIATACode: arrIATA,
		ArrivalAirportTitle:    arrTitle,
		ArrivalAirportCity:     arrCity,
		ArrivalAirportCountry:  arrCountry,
	}, nil
}
