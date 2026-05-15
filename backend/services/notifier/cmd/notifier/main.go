package main

import (
	"io"
	"log/slog"
	"notifier/internal/notifier"
	"os"
	"shared/logger"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

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

	notifier, err := notifier.NewNotifier()
	if err != nil {
		slog.Error("can't start notifier", "err", err)
		os.Exit(1)
	}
	if err := notifier.Run(); err != nil {
		slog.Error("notifier faced error", "err", err)
		os.Exit(1)
	}
	slog.Info("notifier stopped peacefully")
}
