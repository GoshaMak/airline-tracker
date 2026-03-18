package dto

import "airline-tracker/internal/airport/domain"

type GateDTO struct {
	AirportID uint   `json:"airport_id"`
	Number    string `json:"number"`
}

func (g *GateDTO) GateFromDTO() *domain.Gate {
	return &domain.Gate{
		AirportID: g.AirportID,
		Number:    g.Number,
	}
}
