package model

import (
	"api/internal/airport/domain"
	"shared/common"

	"github.com/google/uuid"
)

type AirportModel struct {
	ID       string `bson:"_id"`
	IATACode string `bson:"iata_code"`
	Title    string `bson:"title"`
	City     string `bson:"city"`
	Country  string `bson:"country"`
}

func AirportToModel(a domain.Airport) AirportModel {
	return AirportModel{
		ID:       a.ID.String(),
		IATACode: a.IATACode.String(),
		Title:    a.Title.String(),
		City:     a.City.String(),
		Country:  a.Country.String(),
	}
}

func ModelToAirport(m AirportModel) (domain.Airport, error) {
	uuidID, err := uuid.Parse(m.ID)
	if err != nil {
		return domain.Airport{}, err
	}

	iata, err := domain.NewIATACode(m.IATACode)
	if err != nil {
		return domain.Airport{}, err
	}

	title, err := domain.NewTitle(m.Title)
	if err != nil {
		return domain.Airport{}, err
	}

	city := common.City(m.City)
	country := common.Country(m.Country)

	return domain.Airport{
		ID:       uuidID,
		IATACode: iata,
		Title:    title,
		City:     city,
		Country:  country,
	}, nil
}

type GateModel struct {
	ID        string `bson:"_id"`
	AirportID string `bson:"airport_id"`
	Number    string `bson:"number"`
}

func GateToModel(g domain.Gate) GateModel {
	return GateModel{
		ID:        g.Id.String(),
		AirportID: g.AirportId.String(),
		Number:    g.Number.String(),
	}
}

func ModelToGate(m GateModel) (domain.Gate, error) {
	gateID, err := parseUUID(m.ID)
	if err != nil {
		return domain.Gate{}, err
	}

	airportID, err := parseUUID(m.AirportID)
	if err != nil {
		return domain.Gate{}, err
	}

	number, err := domain.NewGateNumber(m.Number)
	if err != nil {
		return domain.Gate{}, err
	}

	return domain.Gate{
		Id:        gateID,
		AirportId: airportID,
		Number:    number,
	}, nil
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
