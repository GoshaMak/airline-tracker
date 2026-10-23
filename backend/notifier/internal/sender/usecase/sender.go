package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"notifier/internal/receiver/domain/repository"
	"shared/observability"
	"time"

	"github.com/samber/do/v2"
)

type SenderUsecase struct {
	repo          repository.NotificationRepository
	emailSenderUc *EmailSenderUsecase
	workerID      string
}

func NewSenderUsecase(i do.Injector) (*SenderUsecase, error) {
	return &SenderUsecase{
		repo:          do.MustInvoke[repository.NotificationRepository](i),
		emailSenderUc: do.MustInvoke[*EmailSenderUsecase](i),
		workerID:      observability.InstanceID(),
	}, nil
}

const (
	notificationLease        = 30 * time.Second
	notificationCycleTimeout = 10 * time.Second
	maxDeliveryAttempts      = 8
)

func (uc *SenderUsecase) Send(ctx context.Context) error {
	const op = "SenderUsecase.Send"
	ctx, cancel := context.WithTimeout(ctx, notificationCycleTimeout)
	defer cancel()
	notifications, err := uc.repo.ClaimDue(ctx, uc.workerID, 1, notificationLease)
	if err != nil {
		return fmt.Errorf("%s claim: %w", op, err)
	}
	for _, notification := range notifications {
		jobCtx := observability.Extract(ctx, notification.TraceCarrier)
		started := time.Now()
		sendErr := uc.emailSenderUc.SendEmail(jobCtx, notification)
		persistCtx, persistCancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		if sendErr != nil {
			backoff := time.Second << min(notification.Attempts-1, 6)
			backoff += time.Duration(rand.Int63n(int64(250 * time.Millisecond)))
			retryErr := uc.repo.Retry(persistCtx, notification.Id, notification.LeaseToken,
				time.Now().Add(backoff), sendErr.Error(), maxDeliveryAttempts)
			persistCancel()
			result := "retry"
			if notification.Attempts >= maxDeliveryAttempts {
				result = "failed"
			}
			observability.ObserveJob("notifier", result, time.Since(started))
			if retryErr != nil {
				return errors.Join(fmt.Errorf("%s send: %w", op, sendErr), fmt.Errorf("%s persist retry: %w", op, retryErr))
			}
			slog.WarnContext(jobCtx, "email delivery failed",
				"notification_id", notification.Id,
				"attempt", notification.Attempts,
				"err", sendErr)
			continue
		}
		if err := uc.repo.Complete(persistCtx, notification.Id, notification.LeaseToken); err != nil {
			persistCancel()
			observability.ObserveJob("notifier", "retry", time.Since(started))
			return fmt.Errorf("%s complete: %w", op, err)
		}
		persistCancel()
		observability.ObserveJob("notifier", "success", time.Since(started))
		slog.InfoContext(jobCtx, "email delivered",
			"notification_id", notification.Id,
			"attempt", notification.Attempts)
	}
	return nil
}
