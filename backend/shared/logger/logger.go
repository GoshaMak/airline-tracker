package logger

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"shared/observability"

	"go.opentelemetry.io/otel/trace"
)

func SetupLogger(w io.Writer) error {
	service := os.Getenv("SERVICE_NAME")
	if service == "" {
		service = filepath.Base(os.Args[0])
	}
	logger := slog.New(contextHandler{
		Handler: slog.NewJSONHandler(
			w,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		).WithAttrs([]slog.Attr{
			slog.String("service", service),
			slog.String("instance", observability.InstanceID()),
		}),
	})
	if logger == nil {
		return errors.New("failed to setup logger")
	}
	slog.SetDefault(logger)

	return nil
}

type contextHandler struct {
	slog.Handler
}

func (h contextHandler) Handle(ctx context.Context, record slog.Record) error {
	span := trace.SpanContextFromContext(ctx)
	if span.IsValid() {
		record.AddAttrs(slog.String("trace_id", span.TraceID().String()), slog.String("span_id", span.SpanID().String()))
	}
	if id := observability.CorrelationID(ctx); id != "" {
		record.AddAttrs(slog.String("correlation_id", id))
	}
	return h.Handler.Handle(ctx, record)
}

func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{Handler: h.Handler.WithGroup(name)}
}
