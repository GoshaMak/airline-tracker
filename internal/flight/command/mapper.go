package command

import (
	"airline-tracker/internal/flight/domain"
	"airline-tracker/internal/flight/dto"
)

func DTOToFlightCommand(f *dto.FlightDTO) (FlightCommand, error) {
	return FlightCommand{
		ScheduledDeparture: f.ScheduledDeparture,
		ScheduledArrival:   f.ScheduledArrival,
		ActualDeparture:    f.ActualDeparture,
		ActualArrival:      f.ActualArrival,
		Status:             f.Status,
		Plan:               f.Plan,
	}, nil
}

func CommandToFlightDomain(cmd *AddFlightCommand) (domain.Flight, error) {
	f, err := domain.NewFlight(
		cmd.AircraftID,
		cmd.Flight.ScheduledDeparture,
		cmd.Flight.ScheduledArrival,
		cmd.Flight.ActualDeparture,
		cmd.Flight.ActualArrival,
		cmd.Flight.Status,
		cmd.Flight.Plan,
		cmd.DepartureAiroprtID,
		cmd.ArrivalAiroprtID,
		cmd.DepartureGateID,
		cmd.ArrivalGateID,
	)
	return f, err
}
