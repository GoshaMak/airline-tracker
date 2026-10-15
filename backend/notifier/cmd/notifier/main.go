package main

import (
	"context"
	"io"
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
		notifier.Package,
		infra.Package,
		mailer.Package,
		receiver.Package,
		sender.Package,
	)

	notifier, err := do.Invoke[*notifier.Notifier](injector)
	if err != nil {
		slog.Error("can't start notifier", "err", err)
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
	eg.Go(func() error { return notifier.Run(ctx) })
	if err := eg.Wait(); err != nil {
		slog.Error("notifier faced error", "err", err)
		os.Exit(1)
	}
	slog.Info("notifier stopped peacefully")
}
