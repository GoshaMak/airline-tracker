package model

import (
	"api/internal/flight/domain"
	"api/internal/utils"
)

func FlightDomainToModel(f domain.Flight) (Flight, error) {
	var plan *string
	if f.Plan != nil {
		plan = utils.Ptr(f.Plan.String())
	}
	return Flight{
		Id:                 f.Id,
		AircraftId:         f.AircraftId,
		ScheduledDeparture: f.ScheduledDeparture,
		ScheduledArrival:   f.ScheduledArrival,
		ActualDeparture:    f.ActualDeparture,
		ActualArrival:      f.ActualArrival,
		Status:             f.Status.String(),
		Plan:               plan,
		DepartureAirportId: f.DepartureAirportId,
		ArrivalAirportId:   f.ArrivalAirportId,
		DepartureGateId:    f.DepartureGateId,
		ArrivalGateId:      f.ArrivalGateId,
	}, nil
}

func FlightModelToDomain(fm Flight) (domain.Flight, error) {
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
	return domain.Flight{
		Id:                 fm.Id,
		AircraftId:         fm.AircraftId,
		ScheduledDeparture: fm.ScheduledDeparture,
		ScheduledArrival:   fm.ScheduledArrival,
		ActualDeparture:    fm.ActualDeparture,
		ActualArrival:      fm.ActualArrival,
		Status:             st,
		Plan:               p,
		DepartureAirportId: fm.DepartureAirportId,
		ArrivalAirportId:   fm.ArrivalAirportId,
		DepartureGateId:    fm.DepartureGateId,
		ArrivalGateId:      fm.ArrivalGateId,
	}, nil
}
