package model

import (
	"time"

	"github.com/google/uuid"
)

type NotificationModel struct {
	Id        uuid.UUID `db:"id"`
	Payload   []byte    `db:"payload"`
	CreatedAt time.Time `db:"created_at"`
	SendAt    time.Time `db:"send_at"`
	Status    string    `db:"status"`
	Type      string    `db:"type"`
}
