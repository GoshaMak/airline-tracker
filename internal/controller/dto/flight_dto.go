package dto

import (
	"airline-tracker/internal/domain"
	"time"
)

type FlightDTO struct {
	ScheduledDeparture time.Time `json:"scheduled_departure"`
	ScheduledArrival   time.Time `json:"scheduled_arrival"`
	ActualDeparture    time.Time `json:"actual_departure"`
	ActualArrival      time.Time `json:"actual_arrival"`
	Status             string    `json:"status"`
	FlightPlan         string    `json:"flight_plan"`
}

func (f *FlightDTO) FlightFromDTO() *domain.Flight {
	return &domain.Flight{
		ScheduledDeparture: f.ScheduledDeparture,
		ScheduledArrival:   f.ScheduledArrival,
		ActualDeparture:    f.ActualDeparture,
		ActualArrival:      f.ActualArrival,
		Status:             f.Status,
		FlightPlan:         f.FlightPlan,
	}
}

func FlightToDTO(f *domain.Flight) *FlightDTO {
	return &FlightDTO{
		ScheduledDeparture: f.ScheduledDeparture,
		ScheduledArrival:   f.ScheduledArrival,
		ActualDeparture:    f.ActualDeparture,
		ActualArrival:      f.ActualArrival,
		Status:             f.Status,
		FlightPlan:         f.FlightPlan,
	}
}
