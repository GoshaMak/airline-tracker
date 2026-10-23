package main

import (
	"api/internal/infra/kafka"
	"api/internal/infra/postgres"
	"api/internal/publisher"
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"shared/logger"
	"shared/observability"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		slog.Error("publisher failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	if err := logger.SetupLogger(os.Stdout); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	shutdownTracing, err := observability.Init(ctx, "publisher")
	if err != nil {
		return err
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := shutdownTracing(shutdownCtx); err != nil {
			slog.Warn("tracing shutdown failed", "err", err)
		}
	}()

	injector := do.New(
		do.Package(
			do.Lazy(postgres.NewPostgresPool),
			do.Lazy(kafka.NewNotifySender),
		),
		publisher.Package,
	)
	pool, err := do.Invoke[*pgxpool.Pool](injector)
	if err != nil {
		return err
	}
	defer pool.Close()
	sender, err := do.Invoke[*kafka.NotifySender](injector)
	if err != nil {
		return err
	}
	defer func() {
		if err := sender.Close(); err != nil {
			slog.Error("failed to close Kafka producer", "err", err)
		}
	}()

	publisher, err := publisher.NewPublisher(injector)
	if err != nil {
		return err
	}
	slog.Info("publisher started")

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error { return observability.Run(ctx) })
	eg.Go(func() error { return publisher.Run(ctx) })
	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	slog.Info("publisher stopped")

	return nil
}
