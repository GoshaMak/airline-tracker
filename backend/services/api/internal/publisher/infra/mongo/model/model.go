package model

import (
	"api/internal/publisher/domain"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type OutboxDoc struct {
	ID        string     `bson:"_id"`
	Topic     string     `bson:"topic"`
	Payload   bson.Raw   `bson:"payload"`
	CreatedAt time.Time  `bson:"created_at"`
	SentAt    *time.Time `bson:"sent_at,omitempty"`
}

func ToOutboxDoc(ob domain.Outbox) (OutboxDoc, error) {
	payloadBytes, err := bson.Marshal(ob.Payload)
	if err != nil {
		return OutboxDoc{}, err
	}

	return OutboxDoc{
		ID:        ob.Id.String(),
		Topic:     ob.Topic,
		Payload:   payloadBytes,
		CreatedAt: ob.CreatedAt,
		SentAt:    ob.SentAt,
	}, nil
}

func FromOutboxDoc(doc OutboxDoc, targetPayload domain.Payload) (domain.Outbox, error) {
	id, err := uuid.Parse(doc.ID)
	if err != nil {
		return domain.Outbox{}, err
	}

	if err := bson.Unmarshal(doc.Payload, targetPayload); err != nil {
		return domain.Outbox{}, err
	}

	return domain.Outbox{
		Id:        id,
		Topic:     doc.Topic,
		Payload:   targetPayload,
		CreatedAt: doc.CreatedAt,
		SentAt:    doc.SentAt,
	}, nil
}
