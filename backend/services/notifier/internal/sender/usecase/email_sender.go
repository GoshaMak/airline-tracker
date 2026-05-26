package usecase

import (
	"context"
	"fmt"
	"notifier/internal/mailer"
	"notifier/internal/receiver/domain"
	"notifier/internal/sender/command"
	"shared/common"
	"strings"
	"time"

	"github.com/samber/do/v2"
	"golang.org/x/sync/errgroup"
)

type EmailSenderUsecase struct {
	m *mailer.Mailer
}

func NewEmailSenderUsecase(i do.Injector) (*EmailSenderUsecase, error) {
	return &EmailSenderUsecase{
		m: do.MustInvoke[*mailer.Mailer](i),
	}, nil
}

const (
	sep = "\r\n"
)

func (uc *EmailSenderUsecase) SendEmail(ctx context.Context, n domain.Notification) error {
	const op = "EmailSenderUsecase.SendEmail"

	switch n.Type {
	case domain.NotificationSubscribed:
		cmd, err := command.NewSendSubscribedNotificationCommand(n)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		if err := uc.sendSubscribed(ctx, cmd); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case domain.NotificationFlightUpdated:
		cmd, err := command.NewSendFlightUpdatedCommand(n)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		if err := uc.sendFlightUpdated(ctx, cmd); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (uc *EmailSenderUsecase) sendSubscribed(ctx context.Context, cmd command.SendSubscribedNotificationCommand) error {
	var subj, body string
	subj = "Flight departure"
	body = formDepartureBody(cmd)

	msg := []byte(
		"From: " + uc.m.AppEmail.String() + sep +
			"To: " + cmd.ToEmail.String() + sep +
			"Subject: " + subj + sep +
			sep +
			body + sep,
	)

	if err := uc.m.SendEmail([]common.Email{cmd.ToEmail}, msg); err != nil {
		return err
	}

	return nil
}

func formDepartureBody(cmd command.SendSubscribedNotificationCommand) string {
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

func (uc *EmailSenderUsecase) sendFlightUpdated(
	ctx context.Context,
	cmd command.SendFlightUpdatedCommand,
) error {
	var subj, body string
	subj = "Flight departure"
	body = formFlightUpdatedBody(cmd)

	g, ctx := errgroup.WithContext(ctx)
	for i := range cmd.Users {
		g.Go(func() error {
			msg := []byte(
				"From: " + uc.m.AppEmail.String() + sep +
					"To: " + cmd.Users[i].String() + sep +
					"Subject: " + subj + sep +
					sep +
					body + sep,
			)

			if err := uc.m.SendEmail([]common.Email{cmd.Users[i]}, msg); err != nil {
				return err
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

func formFlightUpdatedBody(cmd command.SendFlightUpdatedCommand) string {
	var body strings.Builder
	fmt.Fprintf(&body, "Your flight's information from %s to %s has changed"+sep,
		cmd.DepartureAirportTitle,
		cmd.ArrivalAirportTitle,
	)

	if cmd.ScheduledDeparture != nil {
		fmt.Fprintf(&body, "New scheduled departure is %s"+sep, cmd.ScheduledDeparture.Format(time.ANSIC))
	}
	if cmd.ActualDeparture != nil {
		fmt.Fprintf(&body, "New actual departure is %s"+sep, cmd.ActualDeparture.Format(time.ANSIC))
	}

	if cmd.ScheduledArrival != nil {
		fmt.Fprintf(&body, "New scheduled arrival is %s"+sep, cmd.ScheduledArrival.Format(time.ANSIC))
	}
	if cmd.ActualArrival != nil {
		fmt.Fprintf(&body, "New actual arrival is %s"+sep, cmd.ActualArrival.Format(time.ANSIC))
	}

	if cmd.Status != nil {
		fmt.Fprintf(&body, "New flight's status is %s"+sep, *cmd.Status)
	}

	return body.String()
}
