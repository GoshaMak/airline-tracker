package dto

import (
	"time"
)

type FlightDTO struct {
	ScheduledDeparture time.Time  `json:"scheduled_departure" example:"2026-03-26T14:30:00"`
	ScheduledArrival   time.Time  `json:"scheduled_arrival" example:"2026-05-26T14:30:00"`
	ActualDeparture    *time.Time `json:"actual_departure" example:""`
	ActualArrival      *time.Time `json:"actual_arrival" example:""`
	Status             string     `json:"status" example:"waiting"`
	Plan               *string    `json:"plan" example:"ABCDE A99 ABCDE ABC A66"`
}
