package command

import (
	"api/internal/airport/dto"
	"errors"

	"github.com/google/uuid"
)

type CreateGateCommand struct {
	AirportID  uuid.UUID
	GateNumber string
}

func NewCreateGateCommand(req *dto.CreateGateRequest) (*CreateGateCommand, error) {
	if len(req.Gate.Number) == 0 {
		return nil, errors.New("empty gate number")
	}
	return &CreateGateCommand{
		AirportID:  req.Gate.AirportId,
		GateNumber: req.Gate.Number,
	}, nil
}
