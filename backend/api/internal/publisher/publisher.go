package publisher

import (
	"api/internal/publisher/usecase"
	"context"
	"log/slog"
	"time"

	"github.com/samber/do/v2"
)

type Publisher struct {
	injector do.Injector // TODO: remove if not used
	uc       *usecase.PublisherUsecase
}

func NewPublisher(i *do.RootScope) (*Publisher, error) {
	return &Publisher{
		injector: i,
		uc:       do.MustInvoke[*usecase.PublisherUsecase](i),
	}, nil
}

func (p *Publisher) Run(ctx context.Context) error {
	const op = "Publisher.Run"
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		processed, err := p.uc.Publish(ctx)
		if err != nil {
			slog.ErrorContext(ctx, op, "err", err)
		}
		if processed && err == nil {
			continue
		}
		timer := time.NewTimer(time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
