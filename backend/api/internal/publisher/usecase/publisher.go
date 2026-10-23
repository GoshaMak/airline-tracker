package usecase

import (
	"api/internal/infra/kafka"
	"api/internal/publisher/infra/postgres"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"slices"
	"time"

	"shared/observability"

	"github.com/samber/do/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type messageSender interface {
	SendContext(context.Context, string, string, string, int, map[string]string, []byte) error
}

type PublisherUsecase struct {
	ms   messageSender
	repo postgres.OutboxRepository
}

func NewPublisherUsecase(i do.Injector) (*PublisherUsecase, error) {
	return &PublisherUsecase{
		ms:   do.MustInvoke[*kafka.NotifySender](i),
		repo: do.MustInvoke[postgres.OutboxRepository](i),
	}, nil
}

type SendPayload struct {
	data []byte
}

func (p *SendPayload) MarshalJSON() ([]byte, error) {
	if !json.Valid(p.data) {
		return nil, errors.New("invalid JSON payload")
	}
	return p.data, nil
}

func (p *SendPayload) UnmarshalJSON(data []byte) error {
	if !json.Valid(data) {
		return errors.New("invalid JSON payload")
	}
	p.data = slices.Clone(data)
	return nil
}

const (
	outboxLease        = 30 * time.Second
	publishTimeout     = 10 * time.Second
	maxPublishAttempts = 8
)

func (uc *PublisherUsecase) Publish(ctx context.Context) error {
	const op = "PublisherUsecase.Publish"
	ctx, cancel := context.WithTimeout(ctx, publishTimeout)
	defer cancel()
	items, err := uc.repo.Claim(ctx, 1, outboxLease)
	if err != nil {
		return fmt.Errorf("%s claim: %w", op, err)
	}
	for _, item := range items {
		jobCtx := observability.Extract(ctx, item.TraceCarrier)
		jobCtx, span := observability.Start(
			jobCtx,
			"publisher.kafka.publish",
			trace.WithSpanKind(trace.SpanKindProducer),
			trace.WithAttributes(attribute.String("messaging.message.id", item.Id.String())),
		)
		started := time.Now()
		body, sendErr := item.Payload.MarshalJSON()
		if sendErr == nil {
			carrier := observability.Inject(jobCtx)
			sendErr = uc.ms.SendContext(jobCtx, item.Topic, item.Id.String(), item.Id.String(), 1, carrier, body)
		}
		if sendErr != nil {
			span.RecordError(sendErr)
			span.SetStatus(codes.Error, "Kafka publish failed")
		}
		span.End()
		if sendErr != nil {
			backoff := time.Second << min(item.Attempts-1, 6)
			backoff += time.Duration(rand.Int63n(int64(250 * time.Millisecond)))
			persistCtx, persistCancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
			retryErr := uc.repo.Retry(
				persistCtx,
				item.Id,
				item.LeaseToken,
				time.Now().Add(backoff),
				sendErr.Error(),
				maxPublishAttempts,
			)
			persistCancel()
			result := "retry"
			if item.Attempts >= maxPublishAttempts {
				result = "failed"
			}
			observability.ObserveJob("publisher", result, time.Since(started))
			if retryErr != nil {
				return fmt.Errorf("%s retry: %w (publish: %v)", op, retryErr, sendErr)
			}
			slog.InfoContext(jobCtx, "outbox publish failed",
				"event_id", item.Id,
				"attempt", item.Attempts,
				"err", sendErr)
			continue
		}
		persistCtx, persistCancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		if err := uc.repo.Complete(persistCtx, item.Id, item.LeaseToken); err != nil {
			persistCancel()
			observability.ObserveJob("publisher", "retry", time.Since(started))
			return fmt.Errorf("%s complete: %w", op, err)
		}
		persistCancel()
		observability.ObserveJob("publisher", "success", time.Since(started))
		slog.InfoContext(jobCtx, "outbox event published",
			"event_id", item.Id,
			"attempt", item.Attempts,
			"duration", time.Since(started))
	}
	return nil
}
