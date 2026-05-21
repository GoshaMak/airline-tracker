package publisher

import (
	"api/internal/publisher/usecase"
	"context"
	"log/slog"
	"time"

	"github.com/samber/do/v2"
)

type Publisher struct {
	injector do.Injector
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
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			if err := p.uc.Publish(ctx); err != nil {
				slog.Error(op, "err", err)
			}
		}
	}
}
