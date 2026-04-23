package query

import "airline-tracker/internal/airport/dto"

func QueryToListAirportsResponse(q ListAirportsQuery) dto.ListAirportsResponse {
	resp := dto.ListAirportsResponse{}
	for _, a := range q.Airports {
		resp.Airports = append(resp.Airports, dto.AirportResponse{
			ID: a.ID,
			AirportDTO: dto.AirportDTO{
				IATACode: string(a.IATACode),
				Title:    string(a.Title),
				City:     string(a.City),
				Country:  string(a.Country),
			},
		})
	}
	return resp
}
