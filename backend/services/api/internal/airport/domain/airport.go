package domain

import (
	"shared/common"

	"github.com/google/uuid"
)

type Airport struct {
	ID       uuid.UUID
	IATACode IATACode
	Title    Title
	City     common.City
	Country  common.Country
}

func NewAirport(
	iataCode IATACode,
	title Title,
	city common.City,
	country common.Country,
) (Airport, error) {
	return Airport{
		ID:       uuid.New(),
		IATACode: iataCode,
		Title:    title,
		City:     city,
		Country:  country,
	}, nil
}
