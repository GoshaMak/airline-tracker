package model

import "github.com/google/uuid"

type GateModel struct {
	Id        uuid.UUID `db:"id"`
	AirportId uuid.UUID `db:"airport_id"`
	Number    string    `db:"number"`
}
