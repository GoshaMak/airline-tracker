package command

import (
	"airline-tracker/internal/airport/domain"
	"airline-tracker/internal/airport/dto"

	"github.com/google/uuid"
)

type AddGateCommand struct {
	AirportID  uuid.UUID
	GateNumber domain.GateNumber
}

func NewAddGateCommand(req *dto.CreateGateRequest) (*AddGateCommand, error) {
	num, err := domain.NewGateNumber(req.Gate.Number)
	if err != nil {
		return nil, err
	}
	return &AddGateCommand{
		AirportID:  req.Gate.AirportID,
		GateNumber: num,
	}, nil
}
