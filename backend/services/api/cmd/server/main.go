package main

import (
	"api/internal/server"
	"io"
	"log/slog"
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

	srv, err := server.NewServer()
	if err != nil {
		slog.Error("error creating a server", "err", err)
	}
	if err := srv.Run(); err != nil {
		slog.Info("server execution error", "err", err)
		os.Exit(1)
	}
	slog.Info("finished successfully")
	os.Exit(0)
}
