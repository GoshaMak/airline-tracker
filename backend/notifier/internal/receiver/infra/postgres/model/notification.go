package model

import (
	"time"

	"github.com/google/uuid"
)

type NotificationModel struct {
	Id            uuid.UUID  `db:"id"`
	Payload       []byte     `db:"payload"`
	CreatedAt     time.Time  `db:"created_at"`
	SendAt        time.Time  `db:"send_at"`
	Status        string     `db:"status"`
	Type          string     `db:"type"`
	EventID       *string    `db:"event_id"`
	Recipient     *string    `db:"recipient"`
	Attempts      int        `db:"attempts"`
	LeaseToken    *uuid.UUID `db:"lease_token"`
	NextAttemptAt time.Time  `db:"next_attempt_at"`
	TraceCarrier  []byte     `db:"trace_carrier"`
}
