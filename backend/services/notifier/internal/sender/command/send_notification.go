package command

import (
	"encoding/json"
	receiverCommand "notifier/internal/receiver/command"
	"notifier/internal/receiver/domain"
	"shared/common"
	"time"
)

type SendSubscribedNotificationCommand struct {
	ToEmail common.Email

	Departure time.Time

	FlightStatus        string
	FlightStatusChanged bool

	DepartureAirportIATACode string
	DepartureAirportTitle    string
	DepartureAirportCity     common.City
	DepartureAirportCountry  common.Country

	ArrivalAirportIATACode string
	ArrivalAirportTitle    string
	ArrivalAirportCity     common.City
	ArrivalAirportCountry  common.Country
}

func NewSendSubscribedNotificationCommand(n domain.Notification) (SendSubscribedNotificationCommand, error) {
	var pm receiverCommand.SubscriptionCreatedPayloadModel
	if err := json.Unmarshal(n.Payload, &pm); err != nil {
		return SendSubscribedNotificationCommand{}, err
	}

	email, err := common.NewEmail(pm.Email)
	if err != nil {
		return SendSubscribedNotificationCommand{}, err
	}

	dep := pm.ScheduledDeparture
	if pm.ActualDeparture != nil {
		dep = *pm.ActualDeparture
	}

	depCity, err := common.NewCity(pm.DepartureAirportCity)
	if err != nil {
		return SendSubscribedNotificationCommand{}, err
	}
	depCountry, err := common.NewCountry(pm.DepartureAirportCountry)
	if err != nil {
		return SendSubscribedNotificationCommand{}, err
	}

	arrCity, err := common.NewCity(pm.ArrivalAirportCity)
	if err != nil {
		return SendSubscribedNotificationCommand{}, err
	}
	arrCountry, err := common.NewCountry(pm.ArrivalAirportCountry)
	if err != nil {
		return SendSubscribedNotificationCommand{}, err
	}

	return SendSubscribedNotificationCommand{
		ToEmail: email,

		Departure: dep,

		FlightStatus:        pm.FlightStatus,
		FlightStatusChanged: false,

		DepartureAirportIATACode: pm.DepartureAirportIATACode,
		DepartureAirportTitle:    pm.DepartureAirportTitle,
		DepartureAirportCity:     depCity,
		DepartureAirportCountry:  depCountry,

		ArrivalAirportIATACode: pm.ArrivalAirportIATACode,
		ArrivalAirportTitle:    pm.ArrivalAirportTitle,
		ArrivalAirportCity:     arrCity,
		ArrivalAirportCountry:  arrCountry,
	}, nil
}
