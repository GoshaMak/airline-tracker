package usecase

import (
	"api/internal/infra/kafka"
	"api/internal/publisher/domain"
	"api/internal/publisher/domain/repository"
	"api/internal/utils"
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/samber/do/v2"
	"golang.org/x/sync/errgroup"
)

type PublisherUsecase struct {
	ns   *kafka.NotifySender
	repo repository.OutboxRepository
}

func NewPublisherUsecase(i do.Injector) (*PublisherUsecase, error) {
	return &PublisherUsecase{
		ns:   do.MustInvoke[*kafka.NotifySender](i),
		repo: do.MustInvoke[repository.OutboxRepository](i),
	}, nil
}

type SendPayload struct {
	data []byte
}

func (p *SendPayload) MarshalJSON() ([]byte, error) {
	return p.data, nil
}

func (p *SendPayload) UnmarshalJSON(data []byte) error {
	p.data = slices.Clone(data)
	return nil
}

func (uc *PublisherUsecase) Publish(ctx context.Context) error {
	const op = "PublisherUsecase.Publish"
	obs, err := uc.repo.ListNotSent(ctx, func() domain.Payload {
		return &SendPayload{}
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	g, ctx := errgroup.WithContext(ctx)
	for i, ob := range obs {
		g.Go(func() error {
			msg, err := ob.Payload.MarshalJSON()
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			if err := uc.ns.SendMessage(ob.Topic, msg); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			obs[i].SentAt = utils.Ptr(time.Now())

			if err := uc.repo.MarkAsSent(ctx, obs[i]); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
