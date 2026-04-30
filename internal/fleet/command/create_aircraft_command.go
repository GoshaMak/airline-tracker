package command

import (
	"airline-tracker/internal/fleet/dto"
	"fmt"

	"github.com/google/uuid"
)

type CreateAircraftCommand struct {
	RegistrationNumber string
	AircraftModelID    uuid.UUID
	SerialNumber       string
	Mileage            int
}

func NewCreateAircraftCommand(
	req *dto.CreateAircraftRequest,
) (CreateAircraftCommand, error) {
	reg := req.Aircraft.RegistrationNumber
	if reg == "" {
		return CreateAircraftCommand{}, fmt.Errorf("invalid registration number")
	}
	amID := req.Aircraft.AircraftModelID
	serial := req.Aircraft.SerialNumber
	if serial == "" {
		return CreateAircraftCommand{}, fmt.Errorf("ivalid serial number")
	}
	mileage := req.Aircraft.Mileage
	if mileage < 0 {
		return CreateAircraftCommand{}, fmt.Errorf("invalid mileage")
	}
	return CreateAircraftCommand{
		RegistrationNumber: reg,
		AircraftModelID:    amID,
		SerialNumber:       serial,
		Mileage:            mileage,
	}, nil
}
