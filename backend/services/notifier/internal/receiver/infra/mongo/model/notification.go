package model

import (
	"notifier/internal/receiver/domain"
	"time"

	"github.com/google/uuid"
)

type NotificationDoc struct {
	ID        string    `bson:"_id"`
	Payload   []byte    `bson:"payload"`
	CreatedAt time.Time `bson:"created_at"`
	SendAt    time.Time `bson:"send_at"`
	Status    string    `bson:"status"`
	Type      string    `bson:"type"`
}

func ToNotificationDoc(n domain.Notification) NotificationDoc {
	return NotificationDoc{
		ID:        n.Id.String(),
		Payload:   n.Payload,
		CreatedAt: n.CreatedAt,
		SendAt:    n.SendAt,
		Status:    n.Status.String(),
		Type:      n.Type.String(),
	}
}

func FromNotificationDoc(doc NotificationDoc) (domain.Notification, error) {
	id, err := uuid.Parse(doc.ID)
	if err != nil {
		return domain.Notification{}, err
	}

	status, err := domain.NewNotificationStatus(doc.Status)
	if err != nil {
		return domain.Notification{}, err
	}

	notificationType, err := domain.NewNotificationType(doc.Type)
	if err != nil {
		return domain.Notification{}, err
	}

	return domain.Notification{
		Id:        id,
		Payload:   doc.Payload,
		CreatedAt: doc.CreatedAt,
		SendAt:    doc.SendAt,
		Status:    status,
		Type:      notificationType,
	}, nil
}
