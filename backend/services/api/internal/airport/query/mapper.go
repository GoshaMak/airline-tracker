package query

import "api/internal/airport/dto"

func QueryToListAirportsResponse(q ListAirportsQuery) dto.ListAirportsResponse {
	resp := dto.ListAirportsResponse{}
	for _, a := range q.Airports {
		resp.Airports = append(resp.Airports, dto.AirportResponse{
			ID: a.ID,
			AirportDTO: dto.AirportDTO{
				IATACode: a.IATACode.String(),
				Title:    a.Title.String(),
				City:     a.City.String(),
				Country:  a.Country.String(),
			},
		})
	}
	return resp
}
