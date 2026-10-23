package main

import (
	"context"
	"errors"
	"log/slog"
	"notifier/internal/infra"
	"notifier/internal/mailer"
	"notifier/internal/notifier"
	"notifier/internal/receiver"
	"notifier/internal/sender"
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
		slog.Error("notifier failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	if err := logger.SetupLogger(os.Stdout); err != nil {
		return err
	}
	shutdownTracing, err := observability.Init(context.Background(), "notifier")
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
		notifier.Package,
		infra.Package,
		mailer.Package,
		receiver.Package,
		sender.Package,
	)
	pool, err := do.Invoke[*pgxpool.Pool](injector)
	if err != nil {
		return err
	}
	defer pool.Close()

	notifier, err := do.Invoke[*notifier.Notifier](injector)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error { return observability.Run(ctx) })
	eg.Go(func() error { return notifier.Run(ctx) })
	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	slog.Info("notifier stopped peacefully")
	return nil
}
