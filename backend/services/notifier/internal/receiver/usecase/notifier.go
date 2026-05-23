package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"notifier/internal/mailer"
	"notifier/internal/receiver/command"
	"notifier/internal/receiver/domain"
	"notifier/internal/receiver/domain/repository"
	"time"

	"github.com/samber/do/v2"
)

type NotifierUsecase struct {
	m    *mailer.Mailer
	repo repository.NotificationRepository
}

func NewNotifierUsecase(i do.Injector) (*NotifierUsecase, error) {
	return &NotifierUsecase{
		m:    do.MustInvoke[*mailer.Mailer](i),
		repo: do.MustInvokeAs[repository.NotificationRepository](i),
	}, nil
}

func (uc *NotifierUsecase) SaveNotification(
	ctx context.Context,
	cmd command.SubscriptionCreatedCommand,
) error {
	const op = "NotifierUsecase.SaveNotification"
	sendAt := cmd.ScheduledDeparture
	if cmd.ActualDeparture != nil {
		sendAt = *cmd.ActualDeparture
	}
	if sendAt.Before(time.Now().UTC()) {
		slog.Info(op+": notification already expired", "sendAt", sendAt)
		return nil
	}
	n, err := command.SubscriptionCreatedCommandToDomain(&cmd, sendAt, domain.NotificationCreated)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := uc.repo.Save(ctx, n); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
