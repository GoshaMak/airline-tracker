package domain

import (
	"time"

	"github.com/google/uuid"
)

type Outbox struct {
	Id        uuid.UUID
	Topic     string
	Payload   Payload
	CreatedAt time.Time
	SentAt    *time.Time
}

func NewOutbox(topic string, payload Payload) (Outbox, error) {
	return Outbox{
		Id:        uuid.New(),
		Topic:     topic,
		Payload:   payload,
		CreatedAt: time.Now(),
		SentAt:    nil,
	}, nil
}
