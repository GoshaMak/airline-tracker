package model

import "github.com/google/uuid"

type AircraftModelModel struct {
	Id           uuid.UUID `db:"id"`
	Manufacturer string    `db:"manufacturer"`
	Model        string    `db:"model"`
	Mass         int       `db:"mass"`
	MaxAltitude  int       `db:"max_altitude"`
	MaxSpeed     int       `db:"max_speed"`
}
