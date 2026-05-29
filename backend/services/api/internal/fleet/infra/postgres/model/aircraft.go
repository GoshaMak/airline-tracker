package model

import (
	"github.com/google/uuid"
)

type AircraftModel struct {
	Id                 uuid.UUID `db:"id"`
	AircraftModelId    uuid.UUID `db:"aircraft_model_id"`
	RegistrationNumber string    `db:"registration_number"`
	SerialNumber       string    `db:"serial_number"`
	Mileage            int       `db:"mileage"`
}
