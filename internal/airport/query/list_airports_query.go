package query

import "airline-tracker/internal/airport/domain"

type ListAirportsQuery struct {
	Airports []domain.Airport
}
