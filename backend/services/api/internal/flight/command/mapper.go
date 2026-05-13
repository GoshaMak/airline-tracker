package command

import (
	"api/internal/flight/domain"
	"api/internal/flight/dto"
)

func DTOToFlightCommand(f dto.FlightDTO) (FlightCommand, error) {
	return FlightCommand{
		ScheduledDeparture: f.ScheduledDeparture,
		ScheduledArrival:   f.ScheduledArrival,
		ActualDeparture:    f.ActualDeparture,
		ActualArrival:      f.ActualArrival,
		Status:             f.Status,
		Plan:               f.Plan,
	}, nil
}

func CommandToFlightDomain(cmd CreateFlightCommand) (domain.Flight, error) {
	f, err := domain.NewFlight(
		cmd.AircraftId,
		cmd.Flight.ScheduledDeparture,
		cmd.Flight.ScheduledArrival,
		cmd.Flight.ActualDeparture,
		cmd.Flight.ActualArrival,
		cmd.Flight.Status,
		cmd.Flight.Plan,
		cmd.DepartureAiroprtId,
		cmd.ArrivalAiroprtId,
		cmd.DepartureGateId,
		cmd.ArrivalGateId,
	)
	return f, err
}
