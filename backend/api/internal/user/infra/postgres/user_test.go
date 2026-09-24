package postgres

import (
	"api/internal/testutil/pgxfake"
	"api/internal/user/domain"
	"api/internal/user/domain/repository"
	"context"
	"errors"
	"testing"
	"time"

	allure "github.com/allure-framework/allure-go/commons/gotest"

	"shared/common"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func userMother() domain.User {
	return domain.User{Id: uuid.New(), Email: common.Email("traveler@example.com"),
		PasswordHash: domain.PasswordHashed("stored-hash"), Role: domain.UserRole}
}
func userRows(u domain.User) *pgxfake.Rows {
	return &pgxfake.Rows{Columns: []string{"id", "email", "password_hash", "role"},
		Records: [][]any{{u.Id, u.Email.String(), u.PasswordHash.String(), u.Role.String()}}}
}

func TestUserRepository_SaveUser(t *testing.T) {
	for _, tt := range []struct {
		name    string
		dbErr   error
		wantErr error
	}{
		{"positive", nil, nil}, {"negative duplicate", &pgconn.PgError{Code: "23505"}, repository.ErrUserAlreadyExists},
	} {
		allure.Test(t, tt.name, func(allureContext *allure.Context) {
			t := allureContext.T()

			// Arrange
			db := &pgxfake.DB{ExecErr: tt.dbErr}
			repo := &userRepository{conn: db}
			u := userMother()
			// Act
			err := repo.SaveUser(context.Background(), u)
			// Assert
			if (tt.wantErr == nil && err != nil) || (tt.wantErr != nil && !errors.Is(err, tt.wantErr)) {
				t.Fatal(err)
			}
			if len(db.Calls) != 1 || db.Calls[0].Args[0] != u.Id || db.Calls[0].Args[1] != u.Email.String() {
				t.Fatalf("calls=%+v", db.Calls)
			}
		})
	}
}

func TestUserRepository_GetUser(t *testing.T) {
	allure.Test(t, "positive: maps user", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		u := userMother()
		db := &pgxfake.DB{Rows: userRows(u)}
		repo := &userRepository{conn: db}
		// Act
		got, err := repo.GetUser(context.Background(), u.Email.String())
		// Assert
		if err != nil || got.Id != u.Id || got.Email != u.Email || db.Calls[0].Args[0] != u.Email.String() {
			t.Fatalf("got=%+v, err=%v, calls=%+v", got, err, db.Calls)
		}
	})
	allure.Test(t, "negative: not found", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		repo := &userRepository{conn: &pgxfake.DB{Rows: &pgxfake.Rows{}}}
		// Act
		_, err := repo.GetUser(context.Background(), "absent@example.com")
		// Assert
		if !errors.Is(err, repository.ErrUserNotFound) {
			t.Fatal(err)
		}
	})
}

func TestUserRepository_Exist(t *testing.T) {
	allure.Test(t, "positive: maps user by id", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		u := userMother()
		db := &pgxfake.DB{Rows: userRows(u)}
		repo := &userRepository{conn: db}
		// Act
		got, err := repo.Exist(context.Background(), u.Id)
		// Assert
		if err != nil || got.Id != u.Id || db.Calls[0].Args[0] != u.Id {
			t.Fatalf("got=%+v, err=%v, calls=%+v", got, err, db.Calls)
		}
	})
	allure.Test(t, "negative: not found", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		repo := &userRepository{conn: &pgxfake.DB{Rows: &pgxfake.Rows{}}}
		// Act
		_, err := repo.Exist(context.Background(), uuid.New())
		// Assert
		if !errors.Is(err, repository.ErrUserNotFound) {
			t.Fatal(err)
		}
	})
}

func TestUserRepository_Subscribe(t *testing.T) {
	allure.Test(t, "positive: invokes database function", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		db := &pgxfake.DB{}
		repo := &userRepository{conn: db}
		uid, fid := uuid.New(), uuid.New()
		// Act
		err := repo.Subscribe(context.Background(), uid, fid)
		// Assert
		if err != nil || len(db.Calls) != 1 || db.Calls[0].Args[0] != uid || db.Calls[0].Args[1] != fid {
			t.Fatalf("err=%v, calls=%+v", err, db.Calls)
		}
	})
	allure.Test(t, "negative: duplicate subscription", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		db := &pgxfake.DB{ExecErr: &pgconn.PgError{Code: "23505", ConstraintName: "unique_flight_subscription_per_user"}}
		repo := &userRepository{conn: db}
		// Act
		err := repo.Subscribe(context.Background(), uuid.New(), uuid.New())
		// Assert
		if !errors.Is(err, repository.ErrUserAlreadySubscribed) {
			t.Fatal(err)
		}
	})
}

func TestUserRepository_ListFlights(t *testing.T) {
	allure.Test(t, "positive: maps subscribed flight", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		fid, uid := uuid.New(), uuid.New()
		now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
		db := &pgxfake.DB{Rows: &pgxfake.Rows{
			Columns: []string{"id", "aircraft_id", "scheduled_departure", "scheduled_arrival", "actual_departure", "actual_arrival", "status", "plan", "departure_airport_id", "arrival_airport_id", "departure_gate_id", "arrival_gate_id"},
			Records: [][]any{{fid, uuid.New(), now, now.Add(time.Hour), nil, nil, "scheduled", nil, uuid.New(), uuid.New(), uuid.New(), uuid.New()}},
		}}
		repo := &userRepository{conn: db}
		// Act
		got, err := repo.ListFlights(context.Background(), uid)
		// Assert
		if err != nil || len(got) != 1 || got[0].Id != fid || db.Calls[0].Args[0] != uid {
			t.Fatalf("got=%+v, err=%v, calls=%+v", got, err, db.Calls)
		}
	})
	allure.Test(t, "negative: query failure", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		failure := errors.New("database unavailable")
		repo := &userRepository{conn: &pgxfake.DB{QueryErr: failure}}
		// Act
		_, err := repo.ListFlights(context.Background(), uuid.New())
		// Assert
		if !errors.Is(err, failure) {
			t.Fatal(err)
		}
	})
}
