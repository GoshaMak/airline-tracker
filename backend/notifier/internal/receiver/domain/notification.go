package domain

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	Id            uuid.UUID
	Payload       []byte
	CreatedAt     time.Time
	SendAt        time.Time
	Status        NotificationStatus
	Type          NotificationType
	EventID       string
	Recipient     string
	Attempts      int
	LeaseToken    uuid.UUID
	TraceCarrier  map[string]string
	NextAttemptAt time.Time
}

type EventFailure struct {
	EventID       string
	SchemaVersion int
	Carrier       map[string]string
	Topic         string
	Partition     int32
	Offset        int64
	Payload       []byte
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
