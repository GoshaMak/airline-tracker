package model

import (
	"api/internal/flight/domain"
	"api/internal/utils"
	"fmt"

	"github.com/google/uuid"
)

func FlightDomainToModel(f domain.Flight) (FlightModel, error) {
	var plan *string
	if f.Plan != nil {
		plan = utils.Ptr(f.Plan.String())
	}
	return FlightModel{
		Id:                 f.Id,
		AircraftId:         f.AircraftId.String(),
		ScheduledDeparture: f.ScheduledDeparture,
		ScheduledArrival:   f.ScheduledArrival,
		ActualDeparture:    f.ActualDeparture,
		ActualArrival:      f.ActualArrival,
		Status:             f.Status.String(),
		Plan:               plan,
		DepartureAirportId: f.DepartureAirportId.String(),
		ArrivalAirportId:   f.ArrivalAirportId.String(),
		DepartureGateId:    f.DepartureGateId.String(),
		ArrivalGateId:      f.ArrivalGateId.String(),
	}, nil
}

func FlightModelToDomain(fm FlightModel) (domain.Flight, error) {
	st, err := domain.NewFlightStatus(fm.Status)
	if err != nil {
		return domain.Flight{}, err
	}
	var p *domain.FlightPlan
	if fm.Plan != nil {
		pv, err := domain.NewFlightPlan(*fm.Plan)
		if err != nil {
			return domain.Flight{}, err
		}
		p = &pv
	}
	aid, err := uuid.Parse(fm.AircraftId)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("%w: aircraft id", err)
	}
	depAId, err := uuid.Parse(fm.DepartureAirportId)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("%w: departure airport id", err)
	}
	arrAId, err := uuid.Parse(fm.ArrivalAirportId)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("%w: arrival airport id", err)
	}
	depGId, err := uuid.Parse(fm.DepartureGateId)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("%w: departure gate id", err)
	}
	arrGId, err := uuid.Parse(fm.ArrivalGateId)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("%w: arrival gate id", err)
	}
	return domain.Flight{
		Id:                 fm.Id,
		AircraftId:         aid,
		ScheduledDeparture: fm.ScheduledDeparture,
		ScheduledArrival:   fm.ScheduledArrival,
		ActualDeparture:    fm.ActualDeparture,
		ActualArrival:      fm.ActualArrival,
		Status:             st,
		Plan:               p,
		DepartureAirportId: depAId,
		ArrivalAirportId:   arrAId,
		DepartureGateId:    depGId,
		ArrivalGateId:      arrGId,
	}, nil
}
