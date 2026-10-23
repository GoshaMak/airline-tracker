//go:build integration

package integration

import (
	"api/internal/flight/command"
	"errors"
	"testing"

	allure "github.com/allure-framework/allure-go/commons/gotest"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestFlightUsecase_RejectGateFromAnotherAirport(t *testing.T) {
	allure.Test(t, "integration: invalid route leaves PostgreSQL and Redis unchanged", func(a *allure.Context) {
		t := a.T()
		f := newFixture(t)
		ctx := t.Context()
		countRows := func() [2]int {
			var counts [2]int
			if err := f.pool.
				QueryRow(ctx, "SELECT (SELECT count(*) FROM flights), (SELECT count(*) FROM flight_routes)").
				Scan(&counts[0], &counts[1]); err != nil {
				t.Fatal(err)
			}
			return counts
		}
		before := countRows()
		cachedBefore, err := f.cache.GetFlights(ctx)
		if err != nil {
			t.Fatal(err)
		}

		err = f.uc.CreateFlight(ctx, command.CreateFlightCommand{
			Flight: command.FlightCommand{
				ScheduledDeparture: f.flight.ScheduledDeparture,
				ScheduledArrival:   f.flight.ScheduledArrival,
				Status:             "scheduled",
			},
			AircraftId:         f.flight.AircraftId,
			DepartureAiroprtId: f.flight.DepartureAirportId,
			ArrivalAiroprtId:   f.flight.ArrivalAirportId,
			DepartureGateId:    f.flight.ArrivalGateId,
			ArrivalGateId:      f.flight.ArrivalGateId,
		})

		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "P0001" || pgErr.Message != "departure gate does not exist" {
			t.Fatalf("expected invalid departure gate, got %v", err)
		}
		if after := countRows(); after != before {
			t.Fatalf("partial data persisted: before=%v, after=%v", before, after)
		}
		cachedAfter, err := f.cache.GetFlights(ctx)
		if err != nil || len(cachedAfter) != len(cachedBefore) {
			t.Fatalf("cache changed after rejected operation: before=%d, after=%d, err=%v", len(cachedBefore), len(cachedAfter), err)
		}
	})
}
