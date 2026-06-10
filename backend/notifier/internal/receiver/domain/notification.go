package domain

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	Id        uuid.UUID
	Payload   []byte
	CreatedAt time.Time
	SendAt    time.Time
	Status    NotificationStatus
	Type      NotificationType
}

func NewNotification(
	payload []byte,
	sendAt time.Time,
	status NotificationStatus,
	nt NotificationType,
) (Notification, error) {
	return Notification{
		Id:        uuid.New(),
		Payload:   payload,
		CreatedAt: time.Now(),
		SendAt:    sendAt,
		Status:    status,
		Type:      nt,
	}, nil
}
