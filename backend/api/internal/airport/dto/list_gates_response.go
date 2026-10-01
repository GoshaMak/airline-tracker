package dto

import (
	"api/internal/airport/domain"
	"api/internal/pagination"
)

type ListGatesResponse struct {
	Items      []GateDTO `json:"items"`
	NextCursor *string   `json:"next_cursor"`
	HasMore    bool      `json:"has_more"`
}

func ToResponseListGates(page pagination.Page[domain.Gate]) ListGatesResponse {
	mapped := pagination.Map(page, func(g domain.Gate) GateDTO {
		return GateDTO{
			Id:        g.Id,
			AirportId: g.AirportId,
			Number:    g.Number.String(),
		}
	})
	return ListGatesResponse{
		Items:      mapped.Items,
		NextCursor: mapped.NextCursor,
		HasMore:    mapped.HasMore,
	}
}
