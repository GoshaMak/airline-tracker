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
	"io"
	"log/slog"
	"os"
	"os/signal"
	"shared/logger"
	"shared/observability"
	"syscall"

	"github.com/samber/do/v2"
	"golang.org/x/sync/errgroup"
)

func main() {
	var w io.Writer
	switch os.Getenv("MODE") {
	case "DEBUG":
		w = os.Stdout
	default:
		logFileName := os.Getenv("LOG_FILE")
		logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			slog.Error("can't open log file", "err", err)
			return
		}
		defer logFile.Close()
		w = io.MultiWriter(os.Stdout, logFile)
	}
	if err := logger.SetupLogger(w); err != nil {
		slog.Error("can't setup logger", "err", err)
		return
	}

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
		slog.Error("error creating server", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error { return observability.Run(ctx) })
	eg.Go(func() error { return srv.Run(ctx) })
	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Info("app error", "err", err)
		os.Exit(1)
	}
	slog.Info("finished successfully")
}
