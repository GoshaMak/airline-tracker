package domain

import (
	"errors"

	"github.com/google/uuid"
)

type Aircraft struct {
	ID                 uuid.UUID
	RegistrationNumber string
	AircraftModelID    uuid.UUID
	SerialNumber       string
	Mileage            int
}

func NewAircraft(
	registrationNumber string,
	aircraftModelID uuid.UUID,
	serialNumber string,
	mileage int,
) (*Aircraft, error) {
	if registrationNumber == "" {
		return nil, errors.New("registration number is required")
	}
	if registrationNumber == "" {
		return nil, errors.New("registration number is required")
	}

	return &Aircraft{
		ID:                 uuid.New(),
		RegistrationNumber: registrationNumber,
		AircraftModelID:    aircraftModelID,
		SerialNumber:       serialNumber,
		Mileage:            mileage,
	}, nil
}
