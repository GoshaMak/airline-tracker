package dto

import "api/internal/airport/domain"

type ListGatesResponse struct {
	Gates []GateDTO `json:"gates"`
}

func ToResponseListGates(gs []domain.Gate) ListGatesResponse {
	gates := make([]GateDTO, len(gs))
	for i, g := range gs {
		gates[i] = GateDTO{
			Id:        g.Id,
			AirportId: g.AirportId,
			Number:    g.Number.String(),
		}
	}
	return ListGatesResponse{
		Gates: gates,
	}
}
