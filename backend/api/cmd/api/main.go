package main

import (
	"api/internal/airport"
	"api/internal/auth"
	"api/internal/fleet"
	"api/internal/flight"
	"api/internal/infra"
	"api/internal/publisher"
	"api/internal/server"
	"api/internal/user"
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"shared/logger"
	"shared/observability"
	"syscall"
	"time"

	"github.com/samber/do/v2"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		slog.Error("API failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	if err := logger.SetupLogger(os.Stdout); err != nil {
		return err
	}
	shutdownTracing, err := observability.Init(context.Background(), "api")
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
		auth.Package,
		airport.Package,
		fleet.Package,
		flight.Package,
		user.Package,
		infra.Package,
		publisher.OutboxPackage,
	)

	srv, err := server.NewServer(injector)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error { return observability.Run(ctx) })
	eg.Go(func() error { return srv.Run(ctx) })
	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	slog.Info("finished successfully")
	return nil
}
