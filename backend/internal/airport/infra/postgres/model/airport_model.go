package model

import "github.com/google/uuid"

type AirportModel struct {
	ID       uuid.UUID
	IATACode string
	Title    string
	City     string
	Country  string
}
