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
	"io"
	"log/slog"
	"os"
	"os/signal"
	"shared/logger"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/samber/do/v2"
	"golang.org/x/sync/errgroup"
)

func main() {
	godotenv.Load() // TODO: mb too much. as long as docker loads all the variables

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
		w = logFile
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
		//notification.Package,
		publisher.Package,
	)

	srv, err := server.NewServer(injector)
	if err != nil {
		slog.Error("error creating server", "err", err)
	}

	pub, err := publisher.NewPublisher(injector)
	if err != nil {
		slog.Info("error creating publisher", "err", err)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return srv.Run(ctx)
	})

	g.Go(func() error {
		return pub.Run(ctx)
	})

	if err := g.Wait(); err != nil {
		slog.Info("app error", "err", err)
		os.Exit(1)
	}
	slog.Info("finished successfully")
}
