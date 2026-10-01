// Package pgxfake provides a small in-memory pgx fixture for repository unit tests.
package pgxfake

import (
	"context"
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Call struct {
	SQL  string
	Args []any
}

type DB struct {
	Calls    []Call
	ExecErr  error
	QueryErr error
	Rows     pgx.Rows
	Row      pgx.Row
}

func (db *DB) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	db.Calls = append(db.Calls, Call{sql, args})
	return pgconn.NewCommandTag("OK"), db.ExecErr
}

func (db *DB) Query(_ context.Context, sql string, args ...any) (pgx.Rows, error) {
	db.Calls = append(db.Calls, Call{sql, args})
	return db.Rows, db.QueryErr
}

func (db *DB) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	db.Calls = append(db.Calls, Call{sql, args})
	return db.Row
}

type Row struct {
	Values []any
	Err    error
}

func (r Row) Scan(dest ...any) error {
	if r.Err != nil {
		return r.Err
	}
	for i, d := range dest {
		reflect.ValueOf(d).Elem().Set(reflect.ValueOf(r.Values[i]))
	}
	return nil
}

type Rows struct {
	pgx.Rows
	Columns []string
	Records [][]any
	ReadErr error
	index   int
}

func (r *Rows) Close() {}

func (r *Rows) Err() error { return r.ReadErr }

func (r *Rows) Next() bool {
	if r.index >= len(r.Records) {
		return false
	}
	r.index++
	return true
}

func (r *Rows) FieldDescriptions() []pgconn.FieldDescription {
	f := make([]pgconn.FieldDescription, len(r.Columns))
	for i, name := range r.Columns {
		f[i].Name = name
	}
	return f
}

func (r *Rows) RawValues() [][]byte { return make([][]byte, len(r.Columns)) }

func (r *Rows) Scan(dest ...any) error {
	for i, d := range dest {
		v := reflect.ValueOf(d).Elem()
		if r.Records[r.index-1][i] == nil {
			v.SetZero()
		} else {
			v.Set(reflect.ValueOf(r.Records[r.index-1][i]))
		}
	}
	return nil
}
