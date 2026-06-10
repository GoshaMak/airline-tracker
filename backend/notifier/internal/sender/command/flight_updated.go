package command

import (
	"encoding/json"
	"notifier/internal/receiver/command"
	"notifier/internal/receiver/domain"
	"shared/common"
	"time"

	"github.com/google/uuid"
)

type SendFlightUpdatedCommand struct {
	FlightId uuid.UUID
	Users    []common.Email

	DepartureAirportTitle string
	ArrivalAirportTitle   string

	ScheduledDeparture *time.Time
	ActualDeparture    *time.Time

	ScheduledArrival *time.Time
	ActualArrival    *time.Time

	Status *string
	Plan   *string
}

func NewSendFlightUpdatedCommand(n domain.Notification) (SendFlightUpdatedCommand, error) {
	var pm command.FlightUpdatedPayloadModel
	if err := json.Unmarshal(n.Payload, &pm); err != nil {
		return SendFlightUpdatedCommand{}, err
	}
	emails := make([]common.Email, len(pm.Users))
	for i, u := range pm.Users {
		email, err := common.NewEmail(u)
		if err != nil {
			return SendFlightUpdatedCommand{}, err
		}
		emails[i] = email
	}
	return SendFlightUpdatedCommand{
		FlightId:              pm.FlightId,
		Users:                 emails,
		DepartureAirportTitle: pm.DepartureAirportTitle,
		ArrivalAirportTitle:   pm.ArrivalAirportTitle,
		ScheduledDeparture:    pm.ScheduledDeparture,
		ActualDeparture:       pm.ActualDeparture,
		ScheduledArrival:      pm.ScheduledArrival,
		ActualArrival:         pm.ActualArrival,
		Status:                pm.Status,
		Plan:                  pm.Plan,
	}, nil
}
