package postgres

import (
	"context"
	"errors"
	"notifier/internal/receiver/domain"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// notificationBuilder is a Data Builder for repository fixtures.
type notificationBuilder struct{ value domain.Notification }

func newNotificationBuilder() notificationBuilder {
	return notificationBuilder{domain.Notification{
		Id: uuid.New(), Payload: []byte(`{"email":"a@example.com"}`),
		CreatedAt: time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC),
		SendAt:    time.Date(2026, 9, 24, 13, 0, 0, 0, time.UTC),
		Status:    domain.NotificationCreated, Type: domain.NotificationSubscribed,
	}}
}
func (b notificationBuilder) withStatus(s domain.NotificationStatus) notificationBuilder {
	b.value.Status = s
	return b
}
func (b notificationBuilder) build() domain.Notification { return b.value }

type dbCall struct {
	sql  string
	args []any
}
type fakeNotificationDB struct {
	execErr   error
	queryErr  error
	queryRows pgx.Rows
	calls     []dbCall
}

func (db *fakeNotificationDB) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	db.calls = append(db.calls, dbCall{sql, args})
	return pgconn.NewCommandTag("OK"), db.execErr
}
func (db *fakeNotificationDB) Query(_ context.Context, sql string, args ...any) (pgx.Rows, error) {
	db.calls = append(db.calls, dbCall{sql, args})
	return db.queryRows, db.queryErr
}

// fakeRows implements the part of pgx.Rows used by CollectRows and RowToStructByName.
type fakeRows struct {
	pgx.Rows
	columns []string
	values  [][]any
	index   int
	err     error
}

func (r *fakeRows) Close()     {}
func (r *fakeRows) Err() error { return r.err }
func (r *fakeRows) Next() bool {
	if r.index >= len(r.values) {
		return false
	}
	r.index++
	return true
}
func (r *fakeRows) FieldDescriptions() []pgconn.FieldDescription {
	fields := make([]pgconn.FieldDescription, len(r.columns))
	for i, name := range r.columns {
		fields[i].Name = name
	}
	return fields
}
func (r *fakeRows) RawValues() [][]byte { return make([][]byte, len(r.columns)) }
func (r *fakeRows) Scan(dest ...any) error {
	for i, d := range dest {
		reflect.ValueOf(d).Elem().Set(reflect.ValueOf(r.values[r.index-1][i]))
	}
	return nil
}

func TestNotificationRepository_Save(t *testing.T) {
	for _, tt := range []struct {
		name    string
		dbErr   error
		wantErr bool
	}{
		{"positive", nil, false}, {"negative database error", errors.New("write failed"), true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			db := &fakeNotificationDB{execErr: tt.dbErr}
			repo := &notificationRepository{conn: db}
			n := newNotificationBuilder().build()
			// Act
			err := repo.Save(context.Background(), n)
			// Assert
			if (err != nil) != tt.wantErr || (tt.dbErr != nil && !errors.Is(err, tt.dbErr)) {
				t.Fatalf("err=%v", err)
			}
			if len(db.calls) != 1 || !strings.Contains(db.calls[0].sql, "insert into notifications") || db.calls[0].args[0] != n.Id || db.calls[0].args[4] != "created" {
				t.Fatalf("calls=%+v", db.calls)
			}
		})
	}
}

func TestNotificationRepository_ListNotSent(t *testing.T) {
	t.Run("negative: query failure", func(t *testing.T) {
		// Arrange
		failure := errors.New("database unavailable")
		repo := &notificationRepository{conn: &fakeNotificationDB{queryErr: failure}}
		// Act
		got, err := repo.ListNotSent(context.Background())
		// Assert
		if !errors.Is(err, failure) || got != nil {
			t.Fatalf("got=%v, err=%v", got, err)
		}
	})
	t.Run("positive: converts database row", func(t *testing.T) {
		// Arrange
		n := newNotificationBuilder().build()
		db := &fakeNotificationDB{queryRows: &fakeRows{
			columns: []string{"id", "payload", "created_at", "send_at", "status", "type"},
			values:  [][]any{{n.Id, n.Payload, n.CreatedAt, n.SendAt, "created", "subscribed"}},
		}}
		repo := &notificationRepository{conn: db}
		// Act
		got, err := repo.ListNotSent(context.Background())
		// Assert
		if err != nil || len(got) != 1 || got[0].Id != n.Id || got[0].Status != n.Status {
			t.Fatalf("got=%+v, err=%v", got, err)
		}
		if len(db.calls) != 1 || db.calls[0].args[0] != "sent" {
			t.Fatalf("calls=%+v", db.calls)
		}
	})
	t.Run("negative: invalid stored status", func(t *testing.T) {
		// Arrange
		n := newNotificationBuilder().build()
		db := &fakeNotificationDB{queryRows: &fakeRows{
			columns: []string{"id", "payload", "created_at", "send_at", "status", "type"},
			values:  [][]any{{n.Id, n.Payload, n.CreatedAt, n.SendAt, "bad", "subscribed"}},
		}}
		repo := &notificationRepository{conn: db}
		// Act
		got, err := repo.ListNotSent(context.Background())
		// Assert
		if !errors.Is(err, domain.ErrInvalidStatus) || got != nil {
			t.Fatalf("got=%+v, err=%v", got, err)
		}
	})
}

func TestNotificationRepository_Mark(t *testing.T) {
	for _, tt := range []struct {
		name    string
		dbErr   error
		wantErr bool
	}{
		{"positive", nil, false}, {"negative database error", errors.New("write failed"), true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			db := &fakeNotificationDB{execErr: tt.dbErr}
			repo := &notificationRepository{conn: db}
			n := newNotificationBuilder().withStatus(domain.NotificationUrgent).build()
			// Act
			err := repo.Mark(context.Background(), n, domain.NotificationSent)
			// Assert
			if (err != nil) != tt.wantErr || (tt.dbErr != nil && !errors.Is(err, tt.dbErr)) {
				t.Fatalf("err=%v", err)
			}
			if len(db.calls) != 1 || !strings.Contains(db.calls[0].sql, "update notifications") || db.calls[0].args[0] != domain.NotificationSent || db.calls[0].args[1] != n.Id {
				t.Fatalf("calls=%+v", db.calls)
			}
		})
	}
}
