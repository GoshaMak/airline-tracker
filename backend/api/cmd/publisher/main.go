package main

import (
	"api/internal/infra/kafka"
	"api/internal/infra/postgres"
	"api/internal/publisher"
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"shared/logger"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
)

func main() {
	if err := run(); err != nil {
		slog.Error("publisher failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	var output io.Writer = os.Stdout
	if os.Getenv("MODE") != "DEBUG" {
		file, err := os.OpenFile(os.Getenv("LOG_FILE"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
		output = file
	}
	if err := logger.SetupLogger(output); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	injector := do.New(
		do.Package(do.Lazy(postgres.NewPostgresPool), do.Lazy(kafka.NewNotifySender)),
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

	worker, err := publisher.NewPublisher(injector)
	if err != nil {
		return err
	}
	slog.Info("outbox publisher started")

	if err := worker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	slog.Info("outbox publisher stopped")

	return nil
}
