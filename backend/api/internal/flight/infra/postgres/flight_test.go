package postgres

import (
	"api/internal/flight/domain"
	"api/internal/testutil/pgxfake"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	allure "github.com/allure-framework/allure-go/commons/gotest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type flightBuilder struct{ flight domain.Flight }

func newFlightBuilder() flightBuilder {
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	return flightBuilder{domain.Flight{
		Id: uuid.New(), AircraftId: uuid.New(), ScheduledDeparture: now,
		ScheduledArrival: now.Add(time.Hour), Status: domain.FlightStatus(0),
		DepartureAirportId: uuid.New(), ArrivalAirportId: uuid.New(),
		DepartureGateId: uuid.New(), ArrivalGateId: uuid.New(),
	}}
}
func (b flightBuilder) build() domain.Flight { return b.flight }
func flightRows(f domain.Flight) *pgxfake.Rows {
	return &pgxfake.Rows{
		Columns: []string{"id", "aircraft_id", "scheduled_departure", "scheduled_arrival", "actual_departure", "actual_arrival", "status", "plan", "departure_airport_id", "arrival_airport_id", "departure_gate_id", "arrival_gate_id"},
		Records: [][]any{{f.Id, f.AircraftId, f.ScheduledDeparture, f.ScheduledArrival, nil, nil, "scheduled", nil, f.DepartureAirportId, f.ArrivalAirportId, f.DepartureGateId, f.ArrivalGateId}},
	}
}

func TestPostgresDB_Save(t *testing.T) {
	allure.Test(t, "positive: invokes add_flight", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		f := newFlightBuilder().build()
		db := &pgxfake.DB{}
		repo := &PostgresDB{conn: db}
		// Act
		err := repo.Save(context.Background(), f)
		// Assert
		if err != nil || len(db.Calls) != 1 || !strings.Contains(db.Calls[0].SQL, "add_flight") || db.Calls[0].Args[0] != f.Id {
			t.Fatalf("err=%v, calls=%+v", err, db.Calls)
		}
	})
	allure.Test(t, "negative: database error", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		failure := errors.New("write failed")
		repo := &PostgresDB{conn: &pgxfake.DB{ExecErr: failure}}
		// Act
		err := repo.Save(context.Background(), newFlightBuilder().build())
		// Assert
		if !errors.Is(err, failure) {
			t.Fatal(err)
		}
	})
}

func TestPostgresDB_Exist(t *testing.T) {
	allure.Test(t, "positive: maps flight", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		f := newFlightBuilder().build()
		db := &pgxfake.DB{Rows: flightRows(f)}
		repo := &PostgresDB{conn: db}
		// Act
		got, err := repo.Exist(context.Background(), f.Id)
		// Assert
		if err != nil || got.Id != f.Id || db.Calls[0].Args[0] != f.Id {
			t.Fatalf("got=%+v, err=%v, calls=%+v", got, err, db.Calls)
		}
	})
	allure.Test(t, "negative: flight absent", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		repo := &PostgresDB{conn: &pgxfake.DB{Rows: &pgxfake.Rows{}}}
		// Act
		_, err := repo.Exist(context.Background(), uuid.New())
		// Assert
		if !errors.Is(err, ErrFlightNotFound) {
			t.Fatal(err)
		}
	})
}

func TestPostgresDB_Update(t *testing.T) {
	allure.Test(t, "positive: updates selected field", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		f := newFlightBuilder().build()
		db := &pgxfake.DB{Rows: flightRows(f)}
		repo := &PostgresDB{conn: db}
		status := "boarding"
		// Act
		err := repo.Update(context.Background(), domain.UpdateFlightInfo{FlightId: f.Id, Status: &status})
		// Assert
		if err != nil || len(db.Calls) != 2 || !strings.Contains(db.Calls[1].SQL, "status = $1") || db.Calls[1].Args[1] != f.Id {
			t.Fatalf("err=%v, calls=%+v", err, db.Calls)
		}
	})
	allure.Test(t, "negative: flight absent prevents update", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		db := &pgxfake.DB{Rows: &pgxfake.Rows{}}
		repo := &PostgresDB{conn: db}
		// Act
		err := repo.Update(context.Background(), domain.UpdateFlightInfo{FlightId: uuid.New()})
		// Assert
		if !errors.Is(err, ErrFlightNotFound) || len(db.Calls) != 1 {
			t.Fatalf("err=%v, calls=%+v", err, db.Calls)
		}
	})
}

func TestPostgresDB_ListFlights(t *testing.T) {
	allure.Test(t, "positive: maps result", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		f := newFlightBuilder().build()
		repo := &PostgresDB{conn: &pgxfake.DB{Rows: flightRows(f)}}
		// Act
		got, err := repo.ListFlights(context.Background())
		// Assert
		if err != nil || len(got) != 1 || got[0].Id != f.Id {
			t.Fatalf("got=%+v, err=%v", got, err)
		}
	})
	allure.Test(t, "negative: query failure", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		failure := errors.New("database unavailable")
		repo := &PostgresDB{conn: &pgxfake.DB{QueryErr: failure}}
		// Act
		_, err := repo.ListFlights(context.Background())
		// Assert
		if !errors.Is(err, failure) {
			t.Fatal(err)
		}
	})
}

func TestPostgresDB_GetFlightRoute(t *testing.T) {
	allure.Test(t, "positive: maps route", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		fid, rid, dep, arr := uuid.New(), uuid.New(), uuid.New(), uuid.New()
		db := &pgxfake.DB{Rows: &pgxfake.Rows{
			Columns: []string{"id", "flight_id", "departure_gate_id", "arrival_gate_id"},
			Records: [][]any{{rid, fid, dep, arr}},
		}}
		repo := &PostgresDB{conn: db}
		// Act
		got, err := repo.GetFlightRoute(context.Background(), fid)
		// Assert
		if err != nil || got.Id != rid || got.DepartureGateId != dep {
			t.Fatalf("got=%+v, err=%v", got, err)
		}
	})
	allure.Test(t, "negative: route absent", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		repo := &PostgresDB{conn: &pgxfake.DB{Rows: &pgxfake.Rows{}}}
		// Act
		_, err := repo.GetFlightRoute(context.Background(), uuid.New())
		// Assert
		if !errors.Is(err, ErrFlightRouteNotFound) {
			t.Fatal(err)
		}
	})
}

func TestPostgresDB_ListSubscribers(t *testing.T) {
	allure.Test(t, "positive: maps subscriber", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		uid, fid := uuid.New(), uuid.New()
		db := &pgxfake.DB{Rows: &pgxfake.Rows{
			Columns: []string{"id", "email", "password_hash", "role"},
			Records: [][]any{{uid, "traveler@example.com", "hash", "user"}},
		}}
		repo := &PostgresDB{conn: db}
		// Act
		got, err := repo.ListSubscribers(context.Background(), fid)
		// Assert
		if err != nil || len(got) != 1 || got[0].Id != uid || db.Calls[0].Args[0] != fid {
			t.Fatalf("got=%+v, err=%v, calls=%+v", got, err, db.Calls)
		}
	})
	allure.Test(t, "negative: query failure", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		failure := errors.New("database unavailable")
		repo := &PostgresDB{conn: &pgxfake.DB{QueryErr: failure}}
		// Act
		_, err := repo.ListSubscribers(context.Background(), uuid.New())
		// Assert
		if !errors.Is(err, failure) {
			t.Fatal(err)
		}
	})
}

func TestPostgresDB_GetFlightAirports(t *testing.T) {
	allure.Test(t, "positive: maps both airports", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		fid := uuid.New()
		db := &pgxfake.DB{Rows: &pgxfake.Rows{
			Columns: []string{"departure_airport_id", "departure_airport_iata_code", "departure_airport_title", "departure_airport_city", "departure_airport_country", "arrival_airport_id", "arrival_airport_iata_code", "arrival_airport_title", "arrival_airport_city", "arrival_airport_country"},
			Records: [][]any{{uuid.New(), "SVO", "Sheremetyevo", "Moscow", "RU", uuid.New(), "LED", "Pulkovo", "Saint Petersburg", "RU"}},
		}}
		repo := &PostgresDB{conn: db}
		// Act
		dep, arr, err := repo.GetFlightAirports(context.Background(), fid)
		// Assert
		if err != nil || dep.IATACode.String() != "SVO" || arr.IATACode.String() != "LED" || db.Calls[0].Args[0] != fid {
			t.Fatalf("dep=%+v, arr=%+v, err=%v", dep, arr, err)
		}
	})
	allure.Test(t, "negative: no airports", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		repo := &PostgresDB{conn: &pgxfake.DB{Rows: &pgxfake.Rows{}}}
		// Act
		_, _, err := repo.GetFlightAirports(context.Background(), uuid.New())
		// Assert
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Fatal(err)
		}
	})
}
