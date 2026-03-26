package command

import "airline-tracker/internal/airport/domain"
import genDomain "airline-tracker/internal/domain"

func CommandToGateDomain(cmd *AddGateCommand) (*domain.Gate, error) {
	g, err := domain.NewGate(cmd.AirportID, cmd.GateNumber)
	return g, err
}

func CommandToAirportDomain(cmd *AddAirportCommand) (*domain.Airport, error) {
	code, err := domain.NewIATACode(cmd.IATACode)
	if err != nil {
		return nil, err
	}
	title, err := genDomain.NewTitle(cmd.Title)
	if err != nil {
		return nil, err
	}
	city, err := genDomain.NewCity(cmd.City)
	if err != nil {
		return nil, err
	}
	country, err := genDomain.NewCountry(cmd.Country)
	if err != nil {
		return nil, err
	}
	a, err := domain.NewAirport(code, title, city, country)
	return a, err
}
