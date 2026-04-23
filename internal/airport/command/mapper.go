package command

import (
	"airline-tracker/internal/airport/domain"
	"airline-tracker/internal/common"
)

func CommandToGateDomain(cmd *CreateGateCommand) (*domain.Gate, error) {
	g, err := domain.NewGate(cmd.AirportID, cmd.GateNumber)
	return g, err
}

func CommandToAirportDomain(cmd *CreateAirportCommand) (domain.Airport, error) {
	code, err := domain.NewIATACode(cmd.IATACode)
	if err != nil {
		return domain.Airport{}, err
	}
	title, err := common.NewTitle(cmd.Title)
	if err != nil {
		return domain.Airport{}, err
	}
	city, err := common.NewCity(cmd.City)
	if err != nil {
		return domain.Airport{}, err
	}
	country, err := common.NewCountry(cmd.Country)
	if err != nil {
		return domain.Airport{}, err
	}
	a, err := domain.NewAirport(code, title, city, country)
	if err != nil {
		return domain.Airport{}, err
	}
	return a, nil
}
