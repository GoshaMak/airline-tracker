package dto

import (
	"api/internal/airport/domain"
	"api/internal/pagination"
)

func ToResponseListGates(page pagination.Page[domain.Gate]) pagination.Page[GateDTO] {
	return pagination.Map(page, func(g domain.Gate) GateDTO {
		return GateDTO{
			Id:        g.Id,
			AirportId: g.AirportId,
			Number:    g.Number.String(),
		}
	})
}
