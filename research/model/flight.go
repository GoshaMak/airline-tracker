package model

import (
	"time"

	"github.com/gofrs/uuid"
)

type FlightModel struct {
	Id                 uuid.UUID  `db:"id"`
	AircraftId         uuid.UUID  `db:"aircraft_id"`
	ScheduledDeparture time.Time  `db:"scheduled_departure"`
	ScheduledArrival   time.Time  `db:"scheduled_arrival"`
	ActualDeparture    *time.Time `db:"actual_departure"`
	ActualArrival      *time.Time `db:"actual_arrival"`
	Status             string     `db:"status"`
	Plan               *string    `db:"plan"`
}
