//go:build integration

package integration

import (
	"api/internal/flight/domain"
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
		if err != nil || !reflect.DeepEqual(cached, want) {
			t.Fatalf("updated Redis flight=%+v, want=%+v, err=%v", cached, want, err)
		}
		original, err := f.db.Exist(ctx, seedFlightID)
		if err != nil || original.Status != domain.Scheduled || original.Plan.String() != f.flight.Plan.String() {
			t.Fatalf("original flight changed: %+v, err=%v", original, err)
		}
	})
}
