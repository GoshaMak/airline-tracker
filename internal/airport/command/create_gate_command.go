package command

import (
	"airline-tracker/internal/airport/domain"
	"airline-tracker/internal/airport/dto"

	"github.com/google/uuid"
)

type CreateGateCommand struct {
	AirportID  uuid.UUID
	GateNumber domain.GateNumber
}

func NewCreateGateCommand(req *dto.CreateGateRequest) (*CreateGateCommand, error) {
	num, err := domain.NewGateNumber(req.Gate.Number)
	if err != nil {
		return nil, err
	}

	return &CreateGateCommand{
		AirportID:  req.Gate.AirportID,
		GateNumber: num,
	}, nil
}
