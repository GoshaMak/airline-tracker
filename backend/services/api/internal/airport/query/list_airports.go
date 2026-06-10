package query

import "api/internal/airport/domain"

type ListAirportsQuery struct {
	Airports []domain.Airport
}
