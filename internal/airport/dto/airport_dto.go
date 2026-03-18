package dto

import "airline-tracker/internal/airport/domain"

type AirportDTO struct {
	IATACode string `json:"iata_code"`
	Title    string `json:"title"`
	City     string `json:"city"`
	Country  string `json:"country"`
}

func (a *AirportDTO) AirportFromDTO() *domain.Airport {
	return &domain.Airport{
		IATACode: a.IATACode,
		Title:    a.Title,
		City:     a.City,
		Country:  a.Country,
	}
}
