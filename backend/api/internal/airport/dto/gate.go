package dto

import (
	"github.com/google/uuid"
)

type GateDTO struct {
	Id        uuid.UUID `json:"id"`
	AirportId uuid.UUID `json:"airport_id" example:"add manually"`
	Number    string    `json:"number" example:"A1"`
}
