package logger

import (
	"errors"
	"io"
	"log/slog"
)

func SetupLogger(w io.Writer) error {
	logger := slog.New(
		slog.NewTextHandler(
			w,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		))
	if logger == nil {
		return errors.New("failed to setup logger")
	}
	slog.SetDefault(logger)

	return nil
}
