package postgres

import (
	"api/internal/fleet/domain"
	"api/internal/fleet/domain/repository"
	"api/internal/testutil/pgxfake"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// Fleet object mothers supply valid domain values for all repository cases.
func aircraftMother() domain.Aircraft {
	return domain.Aircraft{Id: uuid.New(), AircraftModelId: uuid.New(),
		RegistrationNumber: domain.RegistrationNumber("RA-12345"),
		SerialNumber:       domain.SerialNumber("SN123"), Mileage: domain.Mileage(100)}
}
func aircraftModelMother() domain.AircraftModel {
	return domain.AircraftModel{Id: uuid.New(), Manufacturer: domain.Manufacturer("Boeing"),
		Model: domain.Model("737"), Mass: domain.AircraftMass(100),
		MaxAltitude: domain.AircraftMaxAltitude(100), MaxSpeed: domain.AircraftMaxSpeed(100)}
}

func TestAircraftRepository_SaveAircraft(t *testing.T) {
	for _, tt := range []struct {
		name    string
		dbErr   error
		wantErr error
	}{
		{"positive", nil, nil}, {"negative duplicate", &pgconn.PgError{Code: "23505"}, repository.ErrAircraftAlreadyExists},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			db := &pgxfake.DB{ExecErr: tt.dbErr}
			repo := &aircraftRepository{conn: db}
			a := aircraftMother()
			// Act
			err := repo.SaveAircraft(context.Background(), a)
			// Assert
			if (tt.wantErr == nil && err != nil) || (tt.wantErr != nil && !errors.Is(err, tt.wantErr)) {
				t.Fatal(err)
			}
			if len(db.Calls) != 1 || db.Calls[0].Args[0] != a.RegistrationNumber.String() {
				t.Fatalf("calls=%+v", db.Calls)
			}
		})
	}
}

func TestAircraftRepository_List(t *testing.T) {
	t.Run("positive: maps aircraft", func(t *testing.T) {
		// Arrange
		a := aircraftMother()
		db := &pgxfake.DB{Rows: &pgxfake.Rows{
			Columns: []string{"id", "aircraft_model_id", "registration_number", "serial_number", "mileage"},
			Records: [][]any{{a.Id, a.AircraftModelId, a.RegistrationNumber.String(), a.SerialNumber.String(), 100}},
		}}
		repo := &aircraftRepository{conn: db}
		// Act
		got, err := repo.List(context.Background())
		// Assert
		if err != nil || len(got) != 1 || got[0].Id != a.Id || got[0].RegistrationNumber != a.RegistrationNumber {
			t.Fatalf("got=%+v, err=%v", got, err)
		}
	})
	t.Run("negative: query failure", func(t *testing.T) {
		// Arrange
		failure := errors.New("database unavailable")
		repo := &aircraftRepository{conn: &pgxfake.DB{QueryErr: failure}}
		// Act
		_, err := repo.List(context.Background())
		// Assert
		if !errors.Is(err, failure) {
			t.Fatal(err)
		}
	})
}

func TestAircraftModelRepository_SaveAircraftModel(t *testing.T) {
	for _, tt := range []struct {
		name    string
		dbErr   error
		wantErr error
	}{
		{"positive", nil, nil}, {"negative duplicate", &pgconn.PgError{Code: "23505"}, repository.ErrAircraftModelAlreadyExists},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			db := &pgxfake.DB{ExecErr: tt.dbErr}
			repo := &aircraftModelRepository{conn: db}
			m := aircraftModelMother()
			// Act
			err := repo.SaveAircraftModel(context.Background(), m)
			// Assert
			if (tt.wantErr == nil && err != nil) || (tt.wantErr != nil && !errors.Is(err, tt.wantErr)) {
				t.Fatal(err)
			}
			if len(db.Calls) != 1 || db.Calls[0].Args[0] != m.Manufacturer.String() {
				t.Fatalf("calls=%+v", db.Calls)
			}
		})
	}
}

func TestAircraftModelRepository_GetAircraftModelById(t *testing.T) {
	t.Run("positive: maps aircraft model", func(t *testing.T) {
		// Arrange
		m := aircraftModelMother()
		db := &pgxfake.DB{Rows: &pgxfake.Rows{
			Columns: []string{"id", "manufacturer", "model", "mass", "max_altitude", "max_speed"},
			Records: [][]any{{m.Id, m.Manufacturer.String(), m.Model.String(), 100, 100, 100}},
		}}
		repo := &aircraftModelRepository{conn: db}
		// Act
		got, err := repo.GetAircraftModelById(context.Background(), m.Id)
		// Assert
		if err != nil || got.Id != m.Id || got.Manufacturer != m.Manufacturer {
			t.Fatalf("got=%+v, err=%v", got, err)
		}
	})
	t.Run("negative: not found", func(t *testing.T) {
		// Arrange
		repo := &aircraftModelRepository{conn: &pgxfake.DB{Rows: &pgxfake.Rows{}}}
		// Act
		_, err := repo.GetAircraftModelById(context.Background(), uuid.New())
		// Assert
		if !errors.Is(err, repository.ErrAircraftModelNotFound) {
			t.Fatal(err)
		}
	})
}
