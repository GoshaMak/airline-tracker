package model

import (
	"api/internal/flight/domain"
	"time"

	"github.com/google/uuid"
)

type UpdateFlightInfoModel struct {
	FlightId uuid.UUID `redis:"-"`

	ScheduledDeparture *time.Time `redis:"scheduled_departure,omitempty"`
	ActualDeparture    *time.Time `redis:"actual_departure,omitempty"`

	ScheduledArrival *time.Time `redis:"scheduled_arrival,omitempty"`
	ActualArrival    *time.Time `redis:"actual_arrival,omitempty"`

	Status *string `redis:"status,omitempty"`
	Plan   *string `redis:"plan,omitempty"`
}

func (m *UpdateFlightInfoModel) ToDomain() (domain.UpdateFlightInfo, error) {
	return domain.NewUpdateFlightInfo(
		m.FlightId,
		m.ScheduledDeparture,
		m.ActualDeparture,
		m.ScheduledArrival,
		m.ActualArrival,
		m.Status,
		m.Plan,
	)
}

func UpdateFlightInfoDomainToModel(ufid domain.UpdateFlightInfo) (UpdateFlightInfoModel, error) {
	return UpdateFlightInfoModel{
		FlightId:           ufid.FlightId,
		ScheduledDeparture: ufid.ScheduledDeparture,
		ActualDeparture:    ufid.ActualDeparture,
		ScheduledArrival:   ufid.ScheduledArrival,
		ActualArrival:      ufid.ActualArrival,
		Status:             ufid.Status,
		Plan:               ufid.Plan,
	}, nil
}
