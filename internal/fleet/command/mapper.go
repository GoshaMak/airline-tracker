package command

import "airline-tracker/internal/fleet/domain"

func ToDomainCreateAircraftModelCommand(
	cmd *CreateAircraftModelCommand,
) (domain.AircraftModel, error) {
	am, err := domain.NewAircraftModel(
		cmd.Manufacturer,
		cmd.Model,
		cmd.Mass,
		cmd.MaxAltitude,
		cmd.MaxSpeed,
	)
	return am, err
}

func ToDomainCreateAircraftCommand(
	cmd *CreateAircraftCommand,
) (domain.Aircraft, error) {
	a, err := domain.NewAircraft(
		cmd.RegistrationNumber,
		cmd.AircraftModelID,
		cmd.SerialNumber,
		cmd.Mileage,
	)
	return a, err
}
