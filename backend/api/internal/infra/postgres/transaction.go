package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type transactionKey struct{}

type Beginner interface {
	Begin(context.Context) (pgx.Tx, error)
}

func WithTransaction(ctx context.Context, pool Beginner, fn func(context.Context) error) error {
	if existing := TxFromContext(ctx); existing != nil {
		return fn(ctx)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	txCtx := context.WithValue(ctx, transactionKey{}, tx)
	if err := fn(txCtx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func TxFromContext(ctx context.Context) pgx.Tx {
	tx, _ := ctx.Value(transactionKey{}).(pgx.Tx)
	return tx
}

type Querier interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func Executor(ctx context.Context, fallback Querier) Querier {
	if tx := TxFromContext(ctx); tx != nil {
		return tx
	}
	return fallback
}
