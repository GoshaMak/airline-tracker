package postgres

import (
	"api/internal/airport/domain"
	"api/internal/airport/domain/repository"
	"api/internal/testutil/pgxfake"
	"context"
	"errors"
	"strings"
	"testing"

	allure "github.com/allure-framework/allure-go/commons/gotest"

	"shared/common"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// airportMother is a complete fixture shared by the repository tests.
func airportMother() domain.Airport {
	return domain.Airport{ID: uuid.New(), IATACode: domain.IATACode("SVO"),
		Title: domain.Title("Sheremetyevo"), City: common.City("Moscow"),
		Country: common.Country("RU")}
}
func airportRows(a domain.Airport) *pgxfake.Rows {
	return &pgxfake.Rows{Columns: []string{"id", "iata_code", "title", "city", "country"},
		Records: [][]any{{a.ID, a.IATACode.String(), a.Title.String(), a.City.String(), a.Country.String()}}}
}

func TestAirportRepository_Save(t *testing.T) {
	allure.Test(t, "positive: resolves city and inserts airport", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		a := airportMother()
		cityID := uuid.New()
		db := &pgxfake.DB{Row: pgxfake.Row{Values: []any{cityID}}}
		repo := &airportRepository{conn: db}
		// Act
		err := repo.Save(context.Background(), a)
		// Assert
		if err != nil || len(db.Calls) != 2 || !strings.Contains(db.Calls[1].SQL, "insert into airports") || db.Calls[1].Args[0] != a.ID || db.Calls[1].Args[3] != cityID {
			t.Fatalf("err=%v, calls=%+v", err, db.Calls)
		}
	})
	allure.Test(t, "negative: unknown city", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		db := &pgxfake.DB{Row: pgxfake.Row{Err: pgx.ErrNoRows}}
		repo := &airportRepository{conn: db}
		// Act
		err := repo.Save(context.Background(), airportMother())
		// Assert
		if !errors.Is(err, pgx.ErrNoRows) || len(db.Calls) != 1 {
			t.Fatalf("err=%v, calls=%+v", err, db.Calls)
		}
	})
	allure.Test(t, "negative: duplicate airport", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		db := &pgxfake.DB{Row: pgxfake.Row{Values: []any{uuid.New()}}, ExecErr: &pgconn.PgError{Code: "23505"}}
		repo := &airportRepository{conn: db}
		// Act
		err := repo.Save(context.Background(), airportMother())
		// Assert
		if !errors.Is(err, repository.ErrAirportAlreadyExists) {
			t.Fatal(err)
		}
	})
}

func TestAirportRepository_ListAirports(t *testing.T) {
	allure.Test(t, "positive: maps row to domain", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		a := airportMother()
		repo := &airportRepository{conn: &pgxfake.DB{Rows: airportRows(a)}}
		// Act
		got, err := repo.ListAirports(context.Background())
		// Assert
		if err != nil || len(got) != 1 || got[0].ID != a.ID || got[0].IATACode != a.IATACode {
			t.Fatalf("got=%+v, err=%v", got, err)
		}
	})
	allure.Test(t, "negative: query failure", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		failure := errors.New("database unavailable")
		repo := &airportRepository{conn: &pgxfake.DB{QueryErr: failure}}
		// Act
		got, err := repo.ListAirports(context.Background())
		// Assert
		if !errors.Is(err, failure) || got != nil {
			t.Fatalf("got=%v, err=%v", got, err)
		}
	})
}

func TestGateRepository_Save(t *testing.T) {
	for _, tt := range []struct {
		name    string
		dbErr   error
		wantErr error
	}{
		{"positive", nil, nil}, {"negative duplicate", &pgconn.PgError{Code: "23505"}, repository.ErrGateAlreadyExists},
	} {
		allure.Test(t, tt.name, func(allureContext *allure.Context) {
			t := allureContext.T()

			// Arrange
			db := &pgxfake.DB{ExecErr: tt.dbErr}
			repo := &gateRepository{conn: db}
			gate := domain.Gate{Id: uuid.New(), AirportId: uuid.New(), Number: domain.GateNumber("A1")}
			// Act
			err := repo.Save(context.Background(), gate)
			// Assert
			if (tt.wantErr == nil && err != nil) || (tt.wantErr != nil && !errors.Is(err, tt.wantErr)) {
				t.Fatalf("err=%v", err)
			}
			if len(db.Calls) != 1 || db.Calls[0].Args[0] != gate.Id || db.Calls[0].Args[2] != "A1" {
				t.Fatalf("calls=%+v", db.Calls)
			}
		})
	}
}

func TestGateRepository_GetAirportByGateId(t *testing.T) {
	allure.Test(t, "positive: maps related airport", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		a := airportMother()
		gateID := uuid.New()
		db := &pgxfake.DB{Rows: airportRows(a)}
		repo := &gateRepository{conn: db}
		// Act
		got, err := repo.GetAirportByGateId(context.Background(), gateID)
		// Assert
		if err != nil || got.ID != a.ID || db.Calls[0].Args[0] != gateID {
			t.Fatalf("got=%+v, err=%v, calls=%+v", got, err, db.Calls)
		}
	})
	allure.Test(t, "negative: gate does not exist", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		repo := &gateRepository{conn: &pgxfake.DB{Rows: &pgxfake.Rows{}}}
		// Act
		_, err := repo.GetAirportByGateId(context.Background(), uuid.New())
		// Assert
		if !errors.Is(err, repository.ErrAirportNotFound) {
			t.Fatal(err)
		}
	})
}

func TestGateRepository_List(t *testing.T) {
	allure.Test(t, "positive: maps gate", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		id, airportID := uuid.New(), uuid.New()
		db := &pgxfake.DB{Rows: &pgxfake.Rows{
			Columns: []string{"id", "airport_id", "number"}, Records: [][]any{{id, airportID, "A1"}},
		}}
		repo := &gateRepository{conn: db}
		// Act
		got, err := repo.List(context.Background())
		// Assert
		if err != nil || len(got) != 1 || got[0].Id != id || got[0].Number.String() != "A1" {
			t.Fatalf("got=%+v, err=%v", got, err)
		}
	})
	allure.Test(t, "negative: query failure", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		failure := errors.New("database unavailable")
		repo := &gateRepository{conn: &pgxfake.DB{QueryErr: failure}}
		// Act
		_, err := repo.List(context.Background())
		// Assert
		if !errors.Is(err, failure) {
			t.Fatal(err)
		}
	})
}
