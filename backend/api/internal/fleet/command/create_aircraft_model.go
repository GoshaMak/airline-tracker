package command

import (
	"api/internal/fleet/dto"
	"fmt"
)

type CreateAircraftModelCommand struct {
	Manufacturer string
	Model        string
	Mass         int
	MaxAltitude  int
	MaxSpeed     int
}

func NewCreateAircraftModelCommand(
	req *dto.CreateAircraftModelRequest,
) (CreateAircraftModelCommand, error) {
	manufacturer := req.AircraftModel.Manufacturer
	if manufacturer == "" {
		return CreateAircraftModelCommand{}, fmt.Errorf("invalid manufacturer")
	}
	model := req.AircraftModel.Model
	if model == "" {
		return CreateAircraftModelCommand{}, fmt.Errorf("invalid model")
	}
	mass := req.AircraftModel.Mass
	if mass <= 0 {
		return CreateAircraftModelCommand{}, fmt.Errorf("invalid mass")
	}
	altitude := req.AircraftModel.MaxAltitude
	if altitude <= 0 {
		return CreateAircraftModelCommand{}, fmt.Errorf("invalid max altitude")
	}
	speed := req.AircraftModel.MaxSpeed
	if speed <= 0 {
		return CreateAircraftModelCommand{}, fmt.Errorf("invalid max speed")
	}
	return CreateAircraftModelCommand{
		Manufacturer: manufacturer,
		Model:        model,
		Mass:         mass,
		MaxAltitude:  altitude,
		MaxSpeed:     speed,
	}, nil
}
