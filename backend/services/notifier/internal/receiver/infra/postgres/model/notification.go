package model

import (
	"time"

	"github.com/google/uuid"
)

type NotificationModel struct {
	Id        uuid.UUID `json:"id"`
	Payload   []byte    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
	SendAt    time.Time `json:"send_at"`
	Status    string    `json:"status"`
}
