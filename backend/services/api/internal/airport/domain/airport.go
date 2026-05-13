package domain

import (
	"api/internal/common"

	"github.com/google/uuid"
)

type Airport struct {
	ID       uuid.UUID
	IATACode IATACode
	Title    common.Title
	City     common.City
	Country  common.Country
}

func NewAirport(
	iataCode IATACode,
	title common.Title,
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
