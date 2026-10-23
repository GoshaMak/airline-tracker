package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"notifier/internal/receiver/command"
	"notifier/internal/receiver/domain"
	"notifier/internal/receiver/dto"
	"notifier/internal/receiver/usecase"
	"strconv"
	"time"

	"shared/observability"

	"github.com/IBM/sarama"
	"github.com/samber/do/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type HandlerFunc func(ctx context.Context, msg *sarama.ConsumerMessage) error

const (
	consumerAttemptTimeout = 5 * time.Second
	failurePersistTimeout  = 2 * time.Second
	maxEventAttempts       = 8
)

type eventMetadata struct {
	id      string
	version int
	carrier map[string]string
}

func metadata(msg *sarama.ConsumerMessage) (eventMetadata, error) {
	result := eventMetadata{
		id:      fmt.Sprintf("legacy:%s:%d:%d", msg.Topic, msg.Partition, msg.Offset),
		version: 1,
		carrier: make(map[string]string),
	}
	for _, header := range msg.Headers {
		key, value := string(header.Key), string(header.Value)
		switch key {
		case "event_id":
			if value != "" {
				result.id = value
			}
		case "schema_version":
			version, err := strconv.Atoi(value)
			if err != nil || version < 1 {
				return result, fmt.Errorf("invalid schema_version header")
			}
			result.version = version
		case "correlation_id":
			result.carrier["x-correlation-id"] = value
		case "traceparent", "tracestate", "x-correlation-id":
			result.carrier[key] = value
		}
	}
	return result, nil
}

func SubscriptionCreatedHandler(i do.Injector) HandlerFunc {
	uc := do.MustInvoke[*usecase.NotifierUsecase](i)
	return func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		return process(ctx, msg, uc, func(ctx context.Context, event eventMetadata) error {
			var req dto.SubscriptionCreatedDTO
			if err := json.Unmarshal(msg.Value, &req); err != nil {
				return permanentError{err}
			}
			cmd, err := command.NewSubscriptionCreatedCommand(&req)
			if err != nil {
				return permanentError{err}
			}
			if event.version != 1 {
				return permanentError{fmt.Errorf("unsupported event schema version %d", event.version)}
			}
			return uc.SaveSubscriptionEvent(ctx, cmd, event.id, event.version, observability.Inject(ctx))
		})
	}
}

func FlightUpdatedHandler(i do.Injector) HandlerFunc {
	uc := do.MustInvoke[*usecase.NotifierUsecase](i)
	return func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		return process(ctx, msg, uc, func(ctx context.Context, event eventMetadata) error {
			var req dto.FlightUpdatedDTO
			if err := json.Unmarshal(msg.Value, &req); err != nil {
				return permanentError{err}
			}
			cmd, err := command.NewFlightUpdatedCommand(&req)
			if err != nil {
				return permanentError{err}
			}
			if event.version != 1 {
				return permanentError{fmt.Errorf("unsupported event schema version %d", event.version)}
			}
			return uc.UpdateFlightEvent(ctx, cmd, event.id, event.version, observability.Inject(ctx))
		})
	}
}

type permanentError struct{ err error }

func (e permanentError) Error() string { return e.err.Error() }
func (e permanentError) Unwrap() error { return e.err }

func process(ctx context.Context, msg *sarama.ConsumerMessage, uc *usecase.NotifierUsecase, fn func(context.Context, eventMetadata) error) error {
	started := time.Now()
	event, metadataErr := metadata(msg)
	if metadataErr != nil {
		event = eventMetadata{
			id:      fmt.Sprintf("legacy:%s:%d:%d", msg.Topic, msg.Partition, msg.Offset),
			version: 1,
			carrier: map[string]string{},
		}
	}
	ctx = observability.Extract(ctx, event.carrier)
	ctx, span := observability.Start(
		ctx,
		"notifier.kafka.consume",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(attribute.String("messaging.message.id", event.id)),
	)
	defer span.End()
	jobCtx, cancel := context.WithTimeout(ctx, consumerAttemptTimeout)
	var err error
	if metadataErr != nil {
		err = permanentError{metadataErr}
	} else {
		err = fn(jobCtx, event)
	}
	cancel()
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "Kafka event processing failed")
		maxAttempts := maxEventAttempts
		var permanent permanentError
		if errors.As(err, &permanent) {
			maxAttempts = 1
		}
		persistCtx, persistCancel := context.WithTimeout(context.WithoutCancel(ctx), failurePersistTimeout)
		failure := domain.EventFailure{
			EventID:       event.id,
			SchemaVersion: event.version,
			Carrier:       event.carrier,
			Topic:         msg.Topic,
			Partition:     msg.Partition,
			Offset:        msg.Offset,
			Payload:       msg.Value,
		}
		exhausted, persistErr := uc.RecordEventFailure(persistCtx, failure, err.Error(), maxAttempts)
		persistCancel()
		if persistErr != nil {
			observability.ObserveJob("notifier", "error", time.Since(started))
			return errors.Join(err, persistErr)
		}
		if exhausted {
			observability.ObserveJob("notifier", "failed", time.Since(started))
			slog.ErrorContext(ctx, "Kafka event moved to failed state",
				"event_id", event.id,
				"topic", msg.Topic,
				"partition", msg.Partition,
				"offset", msg.Offset,
				"err", err)
			return nil
		}
		observability.ObserveJob("notifier", "retry", time.Since(started))
		slog.WarnContext(ctx, "Kafka event processing will retry",
			"event_id", event.id,
			"topic", msg.Topic,
			"partition", msg.Partition,
			"offset", msg.Offset,
			"attempt_error", err)
		return err
	}
	observability.ObserveJob("notifier", "success", time.Since(started))
	slog.InfoContext(ctx, "Kafka event processed",
		"event_id", event.id,
		"topic", msg.Topic,
		"partition", msg.Partition,
		"offset", msg.Offset)
	return nil
}
