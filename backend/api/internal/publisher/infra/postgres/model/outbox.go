package model

import (
	"time"

	"github.com/google/uuid"
)

type OutboxModel struct {
	Id        uuid.UUID  `json:"id" db:"id"`
	Topic     string     `json:"topic" db:"topic"`
	Payload   []byte     `json:"payload" db:"payload"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	SentAt    *time.Time `json:"sent_at" db:"sent_at"`
}
