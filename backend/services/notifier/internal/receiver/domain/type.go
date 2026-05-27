package domain

import "errors"

type NotificationType int

var (
	ErrInvalidType = errors.New("invalid notification type")
)

const (
	NotificationSubscribed NotificationType = iota
	NotificationFlightUpdated
)
const (
	notificationSubscribed    = "subscribed"
	notificationFlightUpdated = "flight_updated"
)

func NewNotificationType(nt string) (NotificationType, error) {
	switch nt {
	case notificationSubscribed:
		return NotificationSubscribed, nil

	case notificationFlightUpdated:
		return NotificationFlightUpdated, nil

	default:
		return -1, ErrInvalidType
	}
}

func (s NotificationType) String() string {
	return [...]string{notificationSubscribed, notificationFlightUpdated}[s]
}
