package model

import (
	"api/internal/airport/domain"
	"shared/common"
)

func AirportModelToDomain(am AirportModel) (domain.Airport, error) {
	iata, err := domain.NewIATACode(am.IATACode)
	if err != nil {
		return domain.Airport{}, err
	}
	title, err := domain.NewTitle(am.Title)
	if err != nil {
		return domain.Airport{}, err
	}
	city, err := common.NewCity(am.City)
	if err != nil {
		return domain.Airport{}, err
	}
	country, err := common.NewCountry(am.Country)
	if err != nil {
		return domain.Airport{}, err
	}
	return domain.Airport{
		ID:       am.ID,
		IATACode: iata,
		Title:    title,
		City:     city,
		Country:  country,
	}, nil
}
