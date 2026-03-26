package domain

import (
	"airline-tracker/internal/domain"

	"github.com/google/uuid"
)

type Airport struct {
	ID       uuid.UUID
	IATACode IATACode
	Title    domain.Title
	City     domain.City
	Country  domain.Country
}

func NewAirport(
	iataCode IATACode,
	title domain.Title,
	city domain.City,
	country domain.Country,
) (*Airport, error) {
	return &Airport{
		ID:       uuid.New(),
		IATACode: iataCode,
		Title:    title,
		City:     city,
		Country:  country,
	}, nil
}
