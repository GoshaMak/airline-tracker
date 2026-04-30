package model

import "airline-tracker/internal/flight/domain"

func FlightModelToDomain(f FlightModel) (domain.Flight, error) {
	st, err := domain.NewFlightStatus(f.Status)
	if err != nil {
		return domain.Flight{}, err
	}
	var p *domain.FlightPlan
	if f.Plan != nil {
		pv, err := domain.NewFlightPlan(*f.Plan)
		if err != nil {
			return domain.Flight{}, err
		}
		p = &pv
	}
	return domain.Flight{
		ID:                 f.ID,
		AircraftID:         f.AircraftID,
		ScheduledDeparture: f.ScheduledDeparture,
		ScheduledArrival:   f.ScheduledArrival,
		ActualDeparture:    f.ActualDeparture,
		ActualArrival:      f.ActualArrival,
		Status:             st,
		Plan:               p,
		DepartureAirportID: f.DepartureAirportID,
		ArrivalAirportID:   f.ArrivalAirportID,
		DepartureGateID:    f.DepartureGateID,
		ArrivalGateID:      f.ArrivalGateID,
	}, nil
}
