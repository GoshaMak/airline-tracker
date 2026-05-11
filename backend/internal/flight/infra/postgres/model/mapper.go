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
		Id:                 f.Id,
		AircraftId:         f.AircraftId,
		ScheduledDeparture: f.ScheduledDeparture,
		ScheduledArrival:   f.ScheduledArrival,
		ActualDeparture:    f.ActualDeparture,
		ActualArrival:      f.ActualArrival,
		Status:             st,
		Plan:               p,
		DepartureAirportId: f.DepartureAirportId,
		ArrivalAirportId:   f.ArrivalAirportId,
		DepartureGateId:    f.DepartureGateId,
		ArrivalGateId:      f.ArrivalGateId,
	}, nil
}

func FlightRouteModelToDomain(rm FlightRouteModel) (domain.FlightRoute, error) {
	return domain.FlightRoute{
		Id:              rm.Id,
		FlightId:        rm.FlightId,
		DepartureGateId: rm.DepartureGateId,
		ArrivalGateId:   rm.ArrivalGateId,
	}, nil
}
