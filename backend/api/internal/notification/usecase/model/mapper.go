package model

import "api/internal/notification/domain"

func NotificationDomainToModel(n domain.Notification) (NotificationModel, error) {
	return NotificationModel{
		Email: n.Email.String(),

		ScheduledDeparture: n.ScheduledDeparture,
		ActualDeparture:    n.ActualDeparture,
		FlightStatus:       n.FlightStatus.String(),

		DepartureAirportIATACode: n.DepartureAirportIATACode.String(),
		DepartureAirportTitle:    n.DepartureAirportTitle.String(),
		DepartureAirportCity:     n.DepartureAirportCity.String(),
		DepartureAirportCountry:  n.DepartureAirportCountry.String(),

		ArrivalAirportIATACode: n.ArrivalAirportIATACode.String(),
		ArrivalAirportTitle:    n.ArrivalAirportTitle.String(),
		ArrivalAirportCity:     n.ArrivalAirportCity.String(),
		ArrivalAirportCountry:  n.ArrivalAirportCountry.String(),
	}, nil
}
