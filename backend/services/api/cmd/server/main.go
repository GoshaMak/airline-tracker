package main

import (
	"api/internal/server"
	"log/slog"
	"os"
)

func main() {
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
