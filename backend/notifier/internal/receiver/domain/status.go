package domain

import "errors"

type NotificationStatus int

var (
	ErrInvalidStatus = errors.New("invalid notification status")
)

const (
	NotificationCreated NotificationStatus = iota
	NotificationUrgent
	NotificationSent
	NotificationFailed
)

const (
	created = "created"
	urgent  = "urgent"
	sent    = "sent"
	failed  = "failed"
)

func NewNotificationStatus(status string) (NotificationStatus, error) {
	switch status {
	case created:
		return NotificationCreated, nil

	case urgent:
		return NotificationUrgent, nil

	case sent:
		return NotificationSent, nil
	case failed:
		return NotificationFailed, nil

	default:
		return -1, ErrInvalidStatus
	}
}

func (s NotificationStatus) String() string {
	return [...]string{created, urgent, sent, failed}[s]
}
