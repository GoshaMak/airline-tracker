//go:build integration

package integration

import (
	"api/internal/flight/domain"
	"api/internal/flight/infra/redis"
	"api/internal/pagination"
	"errors"
	"github.com/google/uuid"
	"reflect"
	"testing"

	allure "github.com/allure-framework/allure-go/commons/gotest"
)

func TestFlightRepository_SaveReadUpdate(t *testing.T) {
	allure.Test(t, "integration: save, read and update a flight in PostgreSQL and Redis", func(a *allure.Context) {
		t := a.T()
		f := newFixture(t)
		ctx := t.Context()

		if err := f.repo.Save(ctx, f.flight); err != nil {
			t.Fatal(err)
		}

		saved, err := f.db.Exist(ctx, f.flight.Id)
		if err != nil || !reflect.DeepEqual(saved, f.flight) {
			t.Fatalf("PostgreSQL flight=%+v, want=%+v, err=%v", saved, f.flight, err)
		}
		cached, err := f.cache.GetFlightById(ctx, f.flight.Id)
		if err != nil || !reflect.DeepEqual(cached, f.flight) {
			t.Fatalf("Redis flight=%+v, want=%+v, err=%v", cached, f.flight, err)
		}

		status, plan := "boarding", "DCT SVO LED"
		if err := f.repo.Update(ctx, domain.UpdateFlightInfo{
			FlightId: f.flight.Id, Status: &status, Plan: &plan,
		}); err != nil {
			t.Fatal(err)
		}

		want := f.flight
		want.Status, want.Plan = domain.Boarding, new(domain.FlightPlan(plan))
		updated, err := f.db.Exist(ctx, f.flight.Id)
		if err != nil || !reflect.DeepEqual(updated, want) {
			t.Fatalf("updated PostgreSQL flight=%+v, want=%+v, err=%v", updated, want, err)
		}
		cached, err = f.cache.GetFlightById(ctx, f.flight.Id)
		if !errors.Is(err, redis.ErrFlightNotFound) {
			t.Fatalf("updated flight should be invalidated, got=%+v err=%v", cached, err)
		}
		read, err := f.repo.Exist(ctx, f.flight.Id)
		if err != nil || !reflect.DeepEqual(read, want) {
			t.Fatalf("read after invalidation=%+v err=%v", read, err)
		}
		original, err := f.db.Exist(ctx, seedFlightID)
		if err != nil || original.Status != domain.Scheduled || original.Plan.String() != f.flight.Plan.String() {
			t.Fatalf("original flight changed: %+v, err=%v", original, err)
		}
	})
}

func TestFlightRepository_Pagination(t *testing.T) {
	f := newFixture(t)
	ctx := t.Context()
	if err := f.db.Save(ctx, f.flight); err != nil {
		t.Fatal(err)
	}
	var cursor *uuid.UUID
	seen := map[uuid.UUID]bool{}
	found := false
	for {
		page, err := f.db.ListFlightsPage(ctx, pagination.Params{Limit: 1, Cursor: cursor})
		if err != nil {
			t.Fatal(err)
		}
		if len(page) == 0 {
			break
		}
		if len(page) > 2 {
			t.Fatalf("fetch limit exceeded: %d", len(page))
		}
		flight := page[0]
		if seen[flight.Id] || (cursor != nil && flight.Id.String() <= cursor.String()) {
			t.Fatalf("invalid page order: %v", flight.Id)
		}
		seen[flight.Id] = true
		if flight.Id == f.flight.Id {
			found = true
			if !reflect.DeepEqual(flight, f.flight) {
				t.Fatalf("flight=%+v want=%+v", flight, f.flight)
			}
		}
		cursor = &flight.Id
	}
	if !found {
		t.Fatal("created flight absent from pages")
	}
}
