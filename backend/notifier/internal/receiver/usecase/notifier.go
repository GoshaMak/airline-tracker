package usecase

import (
	"context"
	"fmt"
	"notifier/internal/mailer"
	"notifier/internal/receiver/command"
	"notifier/internal/receiver/domain"
	"notifier/internal/receiver/domain/repository"
	"shared/common"
	"shared/observability"
	"time"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type NotifierUsecase struct {
	m    *mailer.Mailer
	repo repository.NotificationRepository
}

func NewNotifierUsecase(i do.Injector) (*NotifierUsecase, error) {
	return &NotifierUsecase{
		m:    do.MustInvoke[*mailer.Mailer](i),
		repo: do.MustInvoke[repository.NotificationRepository](i),
	}, nil
}

func (uc *NotifierUsecase) RecordEventFailure(
	ctx context.Context,
	event domain.EventFailure,
	cause string,
	maxAttempts int,
) (bool, error) {
	return uc.repo.RecordEventFailure(ctx, event, cause, maxAttempts)
}

func (uc *NotifierUsecase) SaveNotification(
	ctx context.Context,
	cmd command.SubscriptionCreatedCommand,
) error {
	return uc.SaveSubscriptionEvent(ctx, cmd, "manual:"+uuid.NewString(), 1, observability.Inject(ctx))
}

func (uc *NotifierUsecase) SaveSubscriptionEvent(
	ctx context.Context,
	cmd command.SubscriptionCreatedCommand,
	eventID string,
	version int,
	carrier map[string]string,
) error {
	const op = "NotifierUsecase.SaveSubscriptionEvent"
	sendAt := cmd.ScheduledDeparture
	if cmd.ActualDeparture != nil {
		sendAt = *cmd.ActualDeparture
	}
	sendAt = sendAt.Add(-time.Duration(cmd.NotifyBeforeMinutes) * time.Minute)
	if sendAt.Before(time.Now().UTC()) {
		_, err := uc.repo.ApplyEvent(ctx, eventID, version, carrier, nil)
		return err
	}
	n, err := command.SubscriptionCreatedCommandToDomain(&cmd, sendAt, domain.NotificationCreated)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	n.EventID = eventID
	n.Recipient = cmd.Email.String()
	n.TraceCarrier = carrier
	if _, err := uc.repo.ApplyEvent(ctx, eventID, version, carrier, []domain.Notification{n}); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (uc *NotifierUsecase) UpdateFlight(
	ctx context.Context,
	cmd command.FlightUpdatedCommand,
) error {
	return uc.UpdateFlightEvent(ctx, cmd, "manual:"+uuid.NewString(), 1, observability.Inject(ctx))
}

func (uc *NotifierUsecase) UpdateFlightEvent(
	ctx context.Context,
	cmd command.FlightUpdatedCommand,
	eventID string,
	version int,
	carrier map[string]string,
) error {
	const op = "NotifierUsecase.UpdateFlightEvent"
	if len(cmd.Users) == 0 || (cmd.ScheduledDeparture == nil &&
		cmd.ActualDeparture == nil && cmd.ScheduledArrival == nil &&
		cmd.ActualArrival == nil && cmd.Status == nil && cmd.Plan == nil) {
		_, err := uc.repo.ApplyEvent(ctx, eventID, version, carrier, nil)
		return err
	}

	sendAt := time.Now().UTC()
	items := make([]domain.Notification, 0, len(cmd.Users))
	for _, user := range cmd.Users {
		recipientCommand := cmd
		recipientCommand.Users = []common.Email{user}
		n, err := command.FlightUpdatedCommandToDomain(recipientCommand, sendAt, domain.NotificationUrgent)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		n.EventID, n.Recipient, n.TraceCarrier = eventID, user.String(), carrier
		items = append(items, n)
	}
	if _, err := uc.repo.ApplyEvent(ctx, eventID, version, carrier, items); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
