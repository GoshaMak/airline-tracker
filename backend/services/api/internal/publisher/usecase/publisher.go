package usecase

import (
	"api/internal/infra/kafka"
	"api/internal/publisher/domain/repository"
	"api/internal/publisher/infra/postgres/model"
	"api/internal/utils"
	"context"
	"encoding/json"
	"fmt"
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
		repo: do.MustInvokeAs[repository.OutboxRepository](i),
	}, nil
}

func (uc *PublisherUsecase) Publish(ctx context.Context) error {
	const op = "PublisherUsecase.Publish"
	obs, err := uc.repo.ListNotSent(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	g, ctx := errgroup.WithContext(ctx)
	for i, ob := range obs {
		g.Go(func() error {
			pm, err := model.PayloadDomainToModel(ob.Payload)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			msg, err := json.Marshal(pm)
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
		return err
	}
	return nil
}
