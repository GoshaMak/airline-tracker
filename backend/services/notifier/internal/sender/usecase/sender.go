package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"notifier/internal/mailer"
	receiverDomain "notifier/internal/receiver/domain"
	receiverRepository "notifier/internal/receiver/domain/repository"
	"notifier/internal/sender/command"
	"notifier/utils"
	"shared/common"
	"time"

	"github.com/samber/do/v2"
	"golang.org/x/sync/errgroup"
)

type SenderUsecase struct {
	repo receiverRepository.NotificationRepository
	m    *mailer.Mailer
}

func NewSenderUsecase(i do.Injector) (*SenderUsecase, error) {
	return &SenderUsecase{
		repo: do.MustInvokeAs[receiverRepository.NotificationRepository](i),
		m:    do.MustInvoke[*mailer.Mailer](i),
	}, nil
}

func (uc *SenderUsecase) Send(ctx context.Context) error {
	const op = "SenderUsecase.Send"

	ns, err := uc.repo.ListByStatus(ctx, receiverDomain.NotificationCreated)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	g, ctx := errgroup.WithContext(ctx)
	now := time.Now().UTC()
	for _, n := range ns {
		g.Go(func() error {
			from := now.Add(-15 * time.Minute)
			to := now.Add(15 * time.Minute)
			if !utils.InTimeSpan(from, to, n.SendAt) {
				slog.Debug(op+": send time not reached yet", "id", n.Id)
				return nil
			}

			cmd, err := command.NewSendNotificationCommand(n)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			if err := uc.sendNotification(cmd); err != nil {
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

func (uc *SenderUsecase) sendNotification(cmd command.SendNotificationCommand) error {
	const op = "SenderUsecase.SendNotification"
	const sep = "\r\n"

	var subj, body string
	if cmd.FlightStatusChanged {
		subj = "Flight status changed"
		body = formStatusChanged(cmd)
	} else {
		subj = "Flight departure"
		body = formDepartureBody(cmd)
	}

	msg := []byte(
		"From: " + uc.m.AppEmail.String() + sep +
			"To: " + cmd.ToEmail.String() + sep +
			"Subject: " + subj + sep +
			sep +
			body + sep,
	)

	if err := uc.m.SendEmail([]common.Email{cmd.ToEmail}, msg); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func formDepartureBody(cmd command.SendNotificationCommand) string {
	const sep = "\r\n"
	body := fmt.Sprintf(
		"Your flight departs at: %s"+sep+
			"From airport %s in %s %s"+sep+
			"Land airport %s in %s %s",
		cmd.Departure.Format(time.ANSIC),

		cmd.DepartureAirportTitle,
		cmd.DepartureAirportCity,
		cmd.DepartureAirportCountry,

		cmd.ArrivalAirportTitle,
		cmd.ArrivalAirportCity,
		cmd.ArrivalAirportCountry,
	)

	return body
}
func formStatusChanged(cmd command.SendNotificationCommand) string {
	panic("sad")
}
