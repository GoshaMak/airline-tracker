package model

import (
	airportDomain "api/internal/airport/domain"
	"api/internal/flight/domain"
	"shared/common"
)

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

func FlightAirportsModelToDomain(
	fam FlightAirportsModel,
) (dep airportDomain.Airport, arr airportDomain.Airport, err error) {
	depIata, err := airportDomain.NewIATACode(fam.DepartureAirportIATACode)
	if err != nil {
		return dep, arr, err
	}
	depTitle, err := airportDomain.NewTitle(fam.DepartureAirportTitle)
	if err != nil {
		return dep, arr, err
	}
	depCity, err := common.NewCity(fam.DepartureAirportCity)
	if err != nil {
		return dep, arr, err
	}
	depCountry, err := common.NewCountry(fam.DepartureAirportCountry)
	if err != nil {
		return dep, arr, err
	}
	dep, err = airportDomain.NewAirport(depIata, depTitle, depCity, depCountry)
	if err != nil {
		return dep, arr, err
	}

	arrIata, err := airportDomain.NewIATACode(fam.ArrivalAirportIATACode)
	if err != nil {
		return arr, arr, err
	}
	arrTitle, err := airportDomain.NewTitle(fam.ArrivalAirportTitle)
	if err != nil {
		return arr, arr, err
	}
	arrCity, err := common.NewCity(fam.ArrivalAirportCity)
	if err != nil {
		return arr, arr, err
	}
	arrCountry, err := common.NewCountry(fam.ArrivalAirportCountry)
	if err != nil {
		return arr, arr, err
	}
	arr, err = airportDomain.NewAirport(arrIata, arrTitle, arrCity, arrCountry)
	if err != nil {
		return dep, arr, err
	}

	return dep, arr, nil
}
