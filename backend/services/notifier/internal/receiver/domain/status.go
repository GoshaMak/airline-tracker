package domain

import "errors"

type NotificationStatus int

var (
	ErrInvalidStatus = errors.New("invalid notification status")
)

const (
	NotificationCreated NotificationStatus = iota
	NotificationPending
	NotificationSent
)

const (
	created = "created"
	pending = "pending"
	sent    = "sent"
)

func NewNotificationStatus(status string) (NotificationStatus, error) {
	switch status {
	case created:
		return NotificationCreated, nil

	case pending:
		return NotificationPending, nil

	case sent:
		return NotificationSent, nil

	default:
		return -1, ErrInvalidStatus
	}
}

func (s NotificationStatus) String() string {
	return [...]string{created, pending, sent}[s]
}
