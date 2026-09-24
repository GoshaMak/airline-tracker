package postgres

import (
	"api/internal/publisher/domain"
	"api/internal/publisher/usecase"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type outboxBuilder struct{ item domain.Outbox }

func newOutboxBuilder() outboxBuilder {
	payload := &usecase.SendPayload{}
	if err := payload.UnmarshalJSON([]byte(`{"flight":"updated"}`)); err != nil {
		panic(err)
	}
	return outboxBuilder{domain.Outbox{
		Id: uuid.New(), Topic: "flights", Payload: payload,
		CreatedAt: time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC),
	}}
}
func (b outboxBuilder) withSentAt(t time.Time) outboxBuilder { b.item.SentAt = &t; return b }
func (b outboxBuilder) build() domain.Outbox                 { return b.item }

type outboxCall struct {
	sql  string
	args []any
}
type fakeOutboxDB struct {
	execErr  error
	queryErr error
	rows     pgx.Rows
	calls    []outboxCall
}

func (db *fakeOutboxDB) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	db.calls = append(db.calls, outboxCall{sql, args})
	return pgconn.NewCommandTag("OK"), db.execErr
}
func (db *fakeOutboxDB) Query(_ context.Context, sql string, args ...any) (pgx.Rows, error) {
	db.calls = append(db.calls, outboxCall{sql, args})
	return db.rows, db.queryErr
}

type outboxRows struct {
	pgx.Rows
	columns []string
	values  [][]any
	index   int
	err     error
}

func (r *outboxRows) Close()     {}
func (r *outboxRows) Err() error { return r.err }
func (r *outboxRows) Next() bool {
	if r.index >= len(r.values) {
		return false
	}
	r.index++
	return true
}
func (r *outboxRows) FieldDescriptions() []pgconn.FieldDescription {
	f := make([]pgconn.FieldDescription, len(r.columns))
	for i, c := range r.columns {
		f[i].Name = c
	}
	return f
}
func (r *outboxRows) RawValues() [][]byte { return make([][]byte, len(r.columns)) }
func (r *outboxRows) Scan(dest ...any) error {
	for i, d := range dest {
		v := reflect.ValueOf(d).Elem()
		if r.values[r.index-1][i] == nil {
			v.SetZero()
		} else {
			v.Set(reflect.ValueOf(r.values[r.index-1][i]))
		}
	}
	return nil
}

type brokenPayload struct{ err error }

func (p *brokenPayload) MarshalJSON() ([]byte, error) { return nil, p.err }
func (p *brokenPayload) UnmarshalJSON([]byte) error   { return p.err }

func TestOutboxRepository_Save(t *testing.T) {
	t.Run("positive: inserts model", func(t *testing.T) {
		// Arrange
		db := &fakeOutboxDB{}
		repo := &outboxRepository{conn: db}
		item := newOutboxBuilder().build()
		// Act
		err := repo.Save(context.Background(), item)
		// Assert
		if err != nil || len(db.calls) != 1 || !strings.Contains(db.calls[0].sql, "insert into outbox") || db.calls[0].args[0] != item.Id {
			t.Fatalf("err=%v, calls=%+v", err, db.calls)
		}
	})
	t.Run("negative: payload encoding fails before insert", func(t *testing.T) {
		// Arrange
		failure := errors.New("encoding failed")
		db := &fakeOutboxDB{}
		repo := &outboxRepository{conn: db}
		item := newOutboxBuilder().build()
		item.Payload = &brokenPayload{err: failure}
		// Act
		err := repo.Save(context.Background(), item)
		// Assert
		if !errors.Is(err, failure) || len(db.calls) != 0 {
			t.Fatalf("err=%v, calls=%+v", err, db.calls)
		}
	})
	t.Run("negative: database error", func(t *testing.T) {
		// Arrange
		failure := errors.New("database unavailable")
		db := &fakeOutboxDB{execErr: failure}
		repo := &outboxRepository{conn: db}
		// Act
		err := repo.Save(context.Background(), newOutboxBuilder().build())
		// Assert
		if !errors.Is(err, failure) {
			t.Fatal(err)
		}
	})
}

func TestOutboxRepository_ListNotSent(t *testing.T) {
	t.Run("negative: query failure", func(t *testing.T) {
		// Arrange
		failure := errors.New("database unavailable")
		repo := &outboxRepository{conn: &fakeOutboxDB{queryErr: failure}}
		// Act
		got, err := repo.ListNotSent(context.Background(), func() domain.Payload { return &usecase.SendPayload{} })
		// Assert
		if !errors.Is(err, failure) || got != nil {
			t.Fatalf("got=%v, err=%v", got, err)
		}
	})
	t.Run("positive: decodes stored payload", func(t *testing.T) {
		// Arrange
		item := newOutboxBuilder().build()
		payload := []byte(`{"flight":"updated"}`)
		db := &fakeOutboxDB{rows: &outboxRows{
			columns: []string{"id", "topic", "payload", "created_at", "sent_at"},
			values:  [][]any{{item.Id, item.Topic, payload, item.CreatedAt, nil}},
		}}
		repo := &outboxRepository{conn: db}
		// Act
		got, err := repo.ListNotSent(context.Background(), func() domain.Payload { return &usecase.SendPayload{} })
		// Assert
		if err != nil || len(got) != 1 || got[0].Id != item.Id {
			t.Fatalf("got=%+v, err=%v", got, err)
		}
		encoded, err := got[0].Payload.MarshalJSON()
		if err != nil || string(encoded) != string(payload) {
			t.Fatalf("payload=%q, err=%v", encoded, err)
		}
	})
	t.Run("negative: payload decoding fails", func(t *testing.T) {
		// Arrange
		failure := errors.New("decode failed")
		item := newOutboxBuilder().build()
		db := &fakeOutboxDB{rows: &outboxRows{
			columns: []string{"id", "topic", "payload", "created_at", "sent_at"},
			values:  [][]any{{item.Id, item.Topic, []byte("broken"), item.CreatedAt, nil}},
		}}
		repo := &outboxRepository{conn: db}
		// Act
		got, err := repo.ListNotSent(context.Background(), func() domain.Payload { return &brokenPayload{err: failure} })
		// Assert
		if !errors.Is(err, failure) || got != nil {
			t.Fatalf("got=%+v, err=%v", got, err)
		}
	})
}

func TestOutboxRepository_MarkAsSent(t *testing.T) {
	for _, tt := range []struct {
		name  string
		dbErr error
	}{
		{"positive", nil}, {"negative database error", errors.New("write failed")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			db := &fakeOutboxDB{execErr: tt.dbErr}
			repo := &outboxRepository{conn: db}
			item := newOutboxBuilder().withSentAt(time.Now()).build()
			// Act
			err := repo.MarkAsSent(context.Background(), item)
			// Assert
			if (err != nil) != (tt.dbErr != nil) || (tt.dbErr != nil && !errors.Is(err, tt.dbErr)) {
				t.Fatalf("err=%v", err)
			}
			if len(db.calls) != 1 || !strings.Contains(db.calls[0].sql, "update outbox") || db.calls[0].args[1] != item.Id {
				t.Fatalf("calls=%+v", db.calls)
			}
		})
	}
}
