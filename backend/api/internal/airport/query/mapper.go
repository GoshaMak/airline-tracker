package query

import (
	"api/internal/airport/domain"
	"api/internal/airport/dto"
	"api/internal/pagination"
)

func QueryToListAirportsResponse(page pagination.Page[domain.Airport]) dto.ListAirportsResponse {
	mapped := pagination.Map(page, func(a domain.Airport) dto.AirportResponse {
		return dto.AirportResponse{
			ID: a.ID,
			AirportDTO: dto.AirportDTO{
				IATACode: a.IATACode.String(),
				Title:    a.Title.String(),
				City:     a.City.String(),
				Country:  a.Country.String(),
			},
		}
	})
	return dto.ListAirportsResponse{
		Items:      mapped.Items,
		NextCursor: mapped.NextCursor,
		HasMore:    mapped.HasMore,
	}
}
