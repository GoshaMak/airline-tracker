//go:build integration

package integration

import (
	"api/internal/flight/command"
	"api/internal/flight/domain"
	flightRedis "api/internal/flight/infra/redis"
	"encoding/json"
	"errors"
	"testing"

	allure "github.com/allure-framework/allure-go/commons/gotest"
	"github.com/google/uuid"
)

func TestFlightUsecase_UpdateCreatesOutboxEvent(t *testing.T) {
	allure.Test(t, "integration: changing a subscribed flight persists an outbox event", func(a *allure.Context) {
		t := a.T()
		f := newFixture(t)
		ctx := t.Context()
		if err := f.repo.Save(ctx, f.flight); err != nil {
			t.Fatal(err)
		}
		if _, err := f.pool.Exec(ctx,
			"INSERT INTO subscriptions(id, user_id, flight_id) VALUES ($1, $2, $3)",
			uuid.New(), adminID, f.flight.Id); err != nil {
			t.Fatal(err)
		}
		if err := f.cache.SaveFlight(ctx, f.flight); err != nil {
			t.Fatalf("warm flight cache: %v", err)
		}

		status := "delayed"
		if err := f.uc.UpdateFlight(ctx, command.UpdateFlightCommand{FlightId: f.flight.Id, Status: &status}); err != nil {
			t.Fatal(err)
		}
		updated, err := f.db.Exist(ctx, f.flight.Id)
		if err != nil || updated.Status != domain.Delayed {
			t.Fatalf("flight status was not persisted: %+v, err=%v", updated, err)
		}
		if _, err := f.cache.GetFlightById(ctx, f.flight.Id); !errors.Is(err, flightRedis.ErrFlightNotFound) {
			t.Fatalf("flight cache was not invalidated after commit: err=%v", err)
		}

		var payload []byte
		var pending bool
		if err := f.pool.QueryRow(ctx,
			"SELECT payload, sent_at IS NULL FROM outbox WHERE topic = $1 AND payload->>'flight_id' = $2",
			"flight-updated", f.flight.Id.String()).Scan(&payload, &pending); err != nil {
			t.Fatalf("outbox event was not committed with the flight update: %v", err)
		}
		var event domain.FlightUpdatedPayloadModel
		if err := json.Unmarshal(payload, &event); err != nil {
			t.Fatal(err)
		}
		if !pending || event.FlightId != f.flight.Id || event.Status == nil ||
			*event.Status != status || len(event.Users) != 1 || event.Users[0] != "admin@airline.test" ||
			event.DepartureAirportTitle != "Sheremetyevo" || event.ArrivalAirportTitle != "Pulkovo" {
			t.Fatalf("unexpected outbox event: %s, pending=%v", payload, pending)
		}

		var eventsBefore int
		if err := f.pool.QueryRow(ctx,
			"SELECT count(*) FROM outbox WHERE topic = $1 AND payload->>'flight_id' = $2",
			"flight-updated", f.flight.Id.String()).Scan(&eventsBefore); err != nil {
			t.Fatal(err)
		}
		if err := f.uc.UpdateFlight(ctx, command.UpdateFlightCommand{FlightId: f.flight.Id, Status: &status}); err != nil {
			t.Fatalf("repeat unchanged update: %v", err)
		}
		var eventsAfter int
		if err := f.pool.QueryRow(ctx,
			"SELECT count(*) FROM outbox WHERE topic = $1 AND payload->>'flight_id' = $2",
			"flight-updated", f.flight.Id.String()).Scan(&eventsAfter); err != nil {
			t.Fatal(err)
		}
		if eventsAfter != eventsBefore {
			t.Fatalf("unchanged update created another outbox event: before=%d after=%d", eventsBefore, eventsAfter)
		}
	})
}
