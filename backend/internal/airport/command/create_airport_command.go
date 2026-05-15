package command

import (
	"airline-tracker/internal/airport/dto"
	"fmt"
)

type CreateAirportCommand struct {
	IATACode string
	Title    string
	City     string
	Country  string
}

func NewCreateAirportCommand(req *dto.CreateAirportRequest) (*CreateAirportCommand, error) {
	code := req.Airport.IATACode
	if code == "" {
		return nil, fmt.Errorf("invalid iata code")
	}
	title := req.Airport.Title
	if title == "" {
		return nil, fmt.Errorf("invalid title")
	}
	city := req.Airport.City
	if city == "" {
		return nil, fmt.Errorf("invalid city")
	}
	country := req.Airport.City
	if country == "" {
		return nil, fmt.Errorf("invalid country")
	}

	return &CreateAirportCommand{
		IATACode: code,
		Title:    title,
		City:     city,
		Country:  country,
	}, nil
}
