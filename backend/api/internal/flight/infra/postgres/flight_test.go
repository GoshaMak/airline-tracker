package postgres

import (
	"api/internal/flight/domain"
	"api/internal/flight/infra/postgres/model"
	"api/internal/pagination"
	"context"
	"errors"
	"reflect"
	"regexp"
	"testing"
	"time"

	allure "github.com/allure-framework/allure-go/commons/gotest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v6"
)

type flightBuilder struct{ flight domain.Flight }

func newFlightBuilder() flightBuilder {
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	return flightBuilder{domain.Flight{
		Id:                 uuid.New(),
		AircraftId:         uuid.New(),
		ScheduledDeparture: now,
		ScheduledArrival:   now.Add(time.Hour),
		Status:             domain.FlightStatus(0),
		DepartureAirportId: uuid.New(),
		ArrivalAirportId:   uuid.New(),
		DepartureGateId:    uuid.New(),
		ArrivalGateId:      uuid.New(),
	}}
}

func (b flightBuilder) withPlan(plan string) flightBuilder {
	b.flight.Plan = new(domain.FlightPlan(plan))
	return b
}

func (b flightBuilder) build() domain.Flight { return b.flight }

func newPostgresFixture(t *testing.T) (*PostgresClient, pgxmock.PgxPoolIface) {
	t.Helper()
	pool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("create pgxmock pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return &PostgresClient{conn: pool}, pool
}

func assertDBExpectations(t *testing.T, pool pgxmock.PgxPoolIface) {
	t.Helper()
	if err := pool.ExpectationsWereMet(); err != nil {
		t.Fatalf("database expectations: %v", err)
	}
}

func flightRows(flights ...domain.Flight) *pgxmock.Rows {
	rowType := reflect.TypeFor[model.FlightModel]()
	columns := make([]string, rowType.NumField())
	for i := range columns {
		columns[i] = rowType.Field(i).Tag.Get("db")
	}
	rows := pgxmock.NewRows(columns)
	for _, f := range flights {
		var plan *string
		if f.Plan != nil {
			value := f.Plan.String()
			plan = &value
		}
		row := reflect.ValueOf(model.FlightModel{
			Id:                 f.Id,
			AircraftId:         f.AircraftId,
			ScheduledDeparture: f.ScheduledDeparture,
			ScheduledArrival:   f.ScheduledArrival,
			ActualDeparture:    f.ActualDeparture,
			ActualArrival:      f.ActualArrival,
			Status:             f.Status.String(),
			Plan:               plan,
			DepartureAirportId: f.DepartureAirportId,
			ArrivalAirportId:   f.ArrivalAirportId,
			DepartureGateId:    f.DepartureGateId,
			ArrivalGateId:      f.ArrivalGateId,
		})
		values := make([]any, row.NumField())
		for i := range values {
			value := row.Field(i)
			if value.Kind() == reflect.Pointer {
				if value.IsNil() {
					continue
				}
				value = value.Elem()
			}
			values[i] = value.Interface()
		}
		rows.AddRow(values...)
	}
	return rows
}

func expectSaveFlight(db pgxmock.PgxPoolIface, f domain.Flight) *pgxmock.ExpectedExec {
	var plan *string
	if f.Plan != nil {
		plan = new(f.Plan.String())
	}
	return db.ExpectExec(regexp.QuoteMeta("select add_flight($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)")).
		WithArgs(f.Id, f.ScheduledDeparture, f.ScheduledArrival, f.Status.String(), plan,
			f.AircraftId, f.DepartureAirportId, f.ArrivalAirportId, f.DepartureGateId, f.ArrivalGateId)
}

func TestPostgresDB_Save(t *testing.T) {
	allure.Test(t, "positive: invokes add_flight", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		f := newFlightBuilder().withPlan("DCT SVO LED").build()
		repo, db := newPostgresFixture(t)
		defer assertDBExpectations(t, db)
		expectSaveFlight(db, f).WillReturnResult(pgxmock.NewResult("SELECT", 1))

		err := repo.Save(context.Background(), f)

		if err != nil {
			t.Fatalf("save flight: %v", err)
		}
	})

	allure.Test(t, "negative: database error", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		failure := errors.New("write failed")
		f := newFlightBuilder().build()
		repo, db := newPostgresFixture(t)
		defer assertDBExpectations(t, db)
		expectSaveFlight(db, f).WillReturnError(failure)

		err := repo.Save(context.Background(), f)

		if !errors.Is(err, failure) {
			t.Fatal(err)
		}
	})
}

func TestPostgresDB_Exist(t *testing.T) {
	allure.Test(t, "positive: maps flight", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		f := newFlightBuilder().build()
		actualDeparture := f.ScheduledDeparture.Add(time.Minute)
		actualArrival := f.ScheduledArrival.Add(time.Minute)
		plan := domain.FlightPlan("DCT SVO LED")
		f.ActualDeparture, f.ActualArrival, f.Plan = &actualDeparture, &actualArrival, &plan
		repo, db := newPostgresFixture(t)
		defer assertDBExpectations(t, db)
		db.ExpectQuery(regexp.QuoteMeta("select * from scan_flight_info($1)")).
			WithArgs(f.Id).WillReturnRows(flightRows(f)).RowsWillBeClosed()

		got, err := repo.Exist(context.Background(), f.Id)

		if err != nil || !reflect.DeepEqual(got, f) {
			t.Fatalf("want=%+v, got=%+v, err=%v", f, got, err)
		}
	})

	allure.Test(t, "negative: flight absent", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		fid := uuid.New()
		repo, db := newPostgresFixture(t)
		defer assertDBExpectations(t, db)
		db.ExpectQuery(regexp.QuoteMeta("select * from scan_flight_info($1)")).
			WithArgs(fid).WillReturnRows(flightRows()).RowsWillBeClosed()

		_, err := repo.Exist(context.Background(), fid)

		if !errors.Is(err, ErrFlightNotFound) {
			t.Fatal(err)
		}
	})
}

func TestPostgresDB_Update(t *testing.T) {
	allure.Test(t, "positive: updates selected field", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		f := newFlightBuilder().build()
		status := "boarding"
		repo, db := newPostgresFixture(t)
		defer assertDBExpectations(t, db)
		db.ExpectQuery(regexp.QuoteMeta("select * from scan_flight_info($1)")).
			WithArgs(f.Id).WillReturnRows(flightRows(f)).RowsWillBeClosed()
		db.ExpectExec(regexp.QuoteMeta("update flights set status = $1 where id = $2")).
			WithArgs(&status, f.Id).WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.Update(context.Background(), domain.UpdateFlightInfo{FlightId: f.Id, Status: &status})

		if err != nil {
			t.Fatalf("update flight: %v", err)
		}
	})

	allure.Test(t, "negative: flight absent prevents update", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		fid := uuid.New()
		status := "boarding"
		repo, db := newPostgresFixture(t)
		defer assertDBExpectations(t, db)
		db.ExpectQuery(regexp.QuoteMeta("select * from scan_flight_info($1)")).
			WithArgs(fid).WillReturnRows(flightRows()).RowsWillBeClosed()

		err := repo.Update(context.Background(), domain.UpdateFlightInfo{FlightId: fid, Status: &status})

		if !errors.Is(err, ErrFlightNotFound) {
			t.Fatalf("want flight not found, got %v", err)
		}
	})
}

func TestPostgresDB_ListFlights(t *testing.T) {
	allure.Test(t, "positive: maps result", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		f := newFlightBuilder().build()
		repo, db := newPostgresFixture(t)
		defer assertDBExpectations(t, db)
		db.ExpectQuery(regexp.QuoteMeta("select * from scan_flights_info()")).
			WithArgs().WillReturnRows(flightRows(f)).RowsWillBeClosed()

		got, err := repo.ListFlights(context.Background())

		if err != nil || !reflect.DeepEqual(got, []domain.Flight{f}) {
			t.Fatalf("got=%+v, err=%v", got, err)
		}
	})

	allure.Test(t, "negative: query failure", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		failure := errors.New("database unavailable")
		repo, db := newPostgresFixture(t)
		defer assertDBExpectations(t, db)
		db.ExpectQuery(regexp.QuoteMeta("select * from scan_flights_info()")).
			WithArgs().WillReturnError(failure)

		_, err := repo.ListFlights(context.Background())

		if !errors.Is(err, failure) {
			t.Fatal(err)
		}
	})
}

func TestPostgresDB_GetFlightRoute(t *testing.T) {
	allure.Test(t, "positive: maps route", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		fid, rid, dep, arr := uuid.New(), uuid.New(), uuid.New(), uuid.New()
		repo, db := newPostgresFixture(t)
		defer assertDBExpectations(t, db)
		rows := pgxmock.NewRows([]string{"id", "flight_id", "departure_gate_id", "arrival_gate_id"}).AddRow(rid, fid, dep, arr)
		db.ExpectQuery(regexp.QuoteMeta("select * from flight_routes where flight_id = $1")).
			WithArgs(fid).WillReturnRows(rows).RowsWillBeClosed()

		got, err := repo.GetFlightRoute(context.Background(), fid)

		want := domain.FlightRoute{Id: rid, FlightId: fid, DepartureGateId: dep, ArrivalGateId: arr}
		if err != nil || got != want {
			t.Fatalf("got=%+v, err=%v", got, err)
		}
	})

	allure.Test(t, "negative: route absent", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		fid := uuid.New()
		repo, db := newPostgresFixture(t)
		defer assertDBExpectations(t, db)
		rows := pgxmock.NewRows([]string{"id", "flight_id", "departure_gate_id", "arrival_gate_id"})
		db.ExpectQuery(regexp.QuoteMeta("select * from flight_routes where flight_id = $1")).
			WithArgs(fid).WillReturnRows(rows).RowsWillBeClosed()

		_, err := repo.GetFlightRoute(context.Background(), fid)

		if !errors.Is(err, ErrFlightRouteNotFound) {
			t.Fatal(err)
		}
	})
}

func TestPostgresDB_ListSubscribers(t *testing.T) {
	const query = "select u.id as id, u.email as email, u.password_hash as password_hash, u.role as role " +
		"from subscriptions s join users u on s.user_id = u.id where s.flight_id = $1::uuid;"

	allure.Test(t, "positive: maps subscriber", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		uid, fid := uuid.New(), uuid.New()
		repo, db := newPostgresFixture(t)
		defer assertDBExpectations(t, db)
		rows := pgxmock.NewRows([]string{"id", "email", "password_hash", "role"}).
			AddRow(uid, "traveler@example.com", "hash", "user")
		db.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(fid).WillReturnRows(rows).RowsWillBeClosed()

		got, err := repo.ListSubscribers(context.Background(), fid)

		if err != nil || len(got) != 1 || got[0].Id != uid || got[0].Email.String() != "traveler@example.com" || got[0].PasswordHash.String() != "hash" || got[0].Role.String() != "user" {
			t.Fatalf("got=%+v, err=%v", got, err)
		}
	})

	allure.Test(t, "negative: query failure", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		failure := errors.New("database unavailable")
		fid := uuid.New()
		repo, db := newPostgresFixture(t)
		defer assertDBExpectations(t, db)
		db.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(fid).WillReturnError(failure)

		_, err := repo.ListSubscribers(context.Background(), fid)

		if !errors.Is(err, failure) {
			t.Fatal(err)
		}
	})
}

func TestPostgresDB_GetFlightAirports(t *testing.T) {
	const query = `(?s)with arps as .*where fr\.flight_id = \$1::uuid.*from arps`
	columns := []string{
		"departure_airport_id", "departure_airport_iata_code", "departure_airport_title",
		"departure_airport_city", "departure_airport_country", "arrival_airport_id",
		"arrival_airport_iata_code", "arrival_airport_title", "arrival_airport_city", "arrival_airport_country",
	}

	allure.Test(t, "positive: maps both airports", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		fid, depID, arrID := uuid.New(), uuid.New(), uuid.New()
		repo, db := newPostgresFixture(t)
		defer assertDBExpectations(t, db)
		rows := pgxmock.NewRows(columns).
			AddRow(depID, "SVO", "Sheremetyevo", "Moscow", "RU", arrID, "LED", "Pulkovo", "Saint Petersburg", "RU")
		db.ExpectQuery(query).WithArgs(fid).WillReturnRows(rows).RowsWillBeClosed()

		dep, arr, err := repo.GetFlightAirports(context.Background(), fid)

		if err != nil || dep.ID != depID || arr.ID != arrID ||
			dep.IATACode.String() != "SVO" || arr.IATACode.String() != "LED" ||
			dep.Title.String() != "Sheremetyevo" || arr.Title.String() != "Pulkovo" ||
			dep.City.String() != "Moscow" || arr.City.String() != "Saint Petersburg" ||
			dep.Country.String() != "RU" || arr.Country.String() != "RU" {
			t.Fatalf("dep=%+v, arr=%+v, err=%v", dep, arr, err)
		}
	})

	allure.Test(t, "negative: no airports", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		fid := uuid.New()
		repo, db := newPostgresFixture(t)
		defer assertDBExpectations(t, db)
		db.ExpectQuery(query).WithArgs(fid).WillReturnRows(pgxmock.NewRows(columns)).RowsWillBeClosed()

		_, _, err := repo.GetFlightAirports(context.Background(), fid)

		if !errors.Is(err, pgx.ErrNoRows) {
			t.Fatal(err)
		}
	})
}

func TestPostgresDB_ListFlightsPage(t *testing.T) {
	const query = "select * from scan_flights_info() where ($1::uuid is null or id > $1) order by id limit $2"
	failure := errors.New("database unavailable")
	for _, tt := range []struct {
		name    string
		cursor  *uuid.UUID
		failure error
	}{
		{
			name:   "positive: forwards cursor and fetch limit",
			cursor: new(uuid.New()),
		},
		{
			name:    "negative: query failure on first page",
			failure: failure,
		},
	} {
		allure.Test(t, tt.name, func(ctx *allure.Context) {
			t := ctx.T()
			params := pagination.Params{Limit: 1, Cursor: tt.cursor}
			flights := []domain.Flight{newFlightBuilder().build(), newFlightBuilder().build()}
			repo, db := newPostgresFixture(t)
			defer assertDBExpectations(t, db)
			var cursor any
			if tt.cursor != nil {
				cursor = *tt.cursor
			}
			expected := db.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(cursor, 2)
			if tt.failure != nil {
				expected.WillReturnError(tt.failure)
			} else {
				expected.WillReturnRows(flightRows(flights...)).RowsWillBeClosed()
			}
			got, err := repo.ListFlightsPage(context.Background(), params)
			if !errors.Is(err, tt.failure) {
				t.Fatalf("err=%v, want=%v", err, tt.failure)
			}
			if tt.failure != nil {
				if got != nil {
					t.Fatalf("flights on error=%+v", got)
				}
				return
			}
			if !reflect.DeepEqual(got, flights) {
				t.Fatalf("got=%+v, want=%+v", got, flights)
			}
		})
	}
}
