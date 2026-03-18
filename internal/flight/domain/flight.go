package domain

import "time"

type Flight struct {
	ID                 uint
	AircraftID         uint
	ScheduledDeparture time.Time
	ScheduledArrival   time.Time
	ActualDeparture    time.Time
	ActualArrival      time.Time
	Status             string
	FlightPlan         string
}
