package domain

import "time"

type Flight struct {
	ID             uint
	ArrivalTime    time.Time
	DepartureTime  time.Time
	ArrivalDelay   time.Time
	DepartureDelay time.Time
}
