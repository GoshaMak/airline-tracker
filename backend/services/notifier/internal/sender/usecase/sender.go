package usecase

import (
	"context"
	"fmt"
	"log/slog"
	receiverDomain "notifier/internal/receiver/domain"
	receiverRepository "notifier/internal/receiver/domain/repository"
	"notifier/utils"
	"time"

	"github.com/samber/do/v2"
	"golang.org/x/sync/errgroup"
)

type SenderUsecase struct {
	repo          receiverRepository.NotificationRepository
	emailSenderUc *EmailSenderUsecase
}

func NewSenderUsecase(i do.Injector) (*SenderUsecase, error) {
	return &SenderUsecase{
		repo:          do.MustInvoke[receiverRepository.NotificationRepository](i),
		emailSenderUc: do.MustInvoke[*EmailSenderUsecase](i),
	}, nil
}

func (uc *SenderUsecase) Send(ctx context.Context) error {
	const op = "SenderUsecase.Send"
	ns, err := uc.repo.ListNotSent(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	g, ctx := errgroup.WithContext(ctx)
	now := time.Now().UTC()
	for _, n := range ns {
		g.Go(func() error {
			if n.Status != receiverDomain.NotificationUrgent {
				from := now.Add(-15 * time.Minute)
				if !utils.InTimeSpan(from, now, n.SendAt) {
					slog.Debug(op+": send time not reached yet", "id", n.Id)
					return nil
				}
			}

			if err := uc.emailSenderUc.SendEmail(ctx, n); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			if err := uc.repo.Mark(ctx, n, receiverDomain.NotificationSent); err != nil {
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
