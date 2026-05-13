package model

import "github.com/google/uuid"

type AirportModel struct {
	ID       uuid.UUID `db:"id"`
	IATACode string    `db:"iata_code"`
	Title    string    `db:"title"`
	City     string    `db:"city"`
	Country  string    `db:"country"`
}
