package domain

import (
	"github.com/google/uuid"
)

type Aircraft struct {
	Id                 uuid.UUID
	RegistrationNumber RegistrationNumber
	AircraftModelId    uuid.UUID
	SerialNumber       SerialNumber
	Mileage            Mileage
}

func NewAircraft(
	registrationNumber string,
	aircraftModelID uuid.UUID,
	serialNumber string,
	mileage int,
) (Aircraft, error) {
	rn, err := NewRegistrationNumber(registrationNumber)
	if err != nil {
		return Aircraft{}, err
	}
	sn, err := NewSerialNumber(serialNumber)
	if err != nil {
		return Aircraft{}, err
	}
	m, err := NewMileage(mileage)
	if err != nil {
		return Aircraft{}, err
	}
	return Aircraft{
		Id:                 uuid.New(),
		RegistrationNumber: rn,
		AircraftModelId:    aircraftModelID,
		SerialNumber:       sn,
		Mileage:            m,
	}, nil
}
