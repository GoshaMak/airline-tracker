package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"notifier/internal/mailer"
	"notifier/internal/receiver/domain"
	"notifier/internal/sender/command"
	"shared/common"
	"strings"
	"time"

	"shared/observability"

	"github.com/samber/do/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type EmailSenderUsecase struct {
	m        mailPort
	appEmail common.Email
}

type mailPort interface {
	SendEmail(context.Context, []common.Email, []byte, string) error
}

func NewEmailSenderUsecase(i do.Injector) (*EmailSenderUsecase, error) {
	m := do.MustInvoke[*mailer.Mailer](i)
	return &EmailSenderUsecase{
		m:        m,
		appEmail: m.AppEmail,
	}, nil
}

const (
	sep = "\r\n"
)

func (uc *EmailSenderUsecase) SendEmail(ctx context.Context, n domain.Notification) error {
	const op = "EmailSenderUsecase.SendEmail"
	ctx, span := observability.Start(
		ctx,
		"notifier.smtp.send",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.String("notification.id", n.Id.String())),
	)
	defer span.End()
	var err error

	switch n.Type {
	case domain.NotificationSubscribed:
		var cmd command.SendSubscribedNotificationCommand
		cmd, err = command.NewSendSubscribedNotificationCommand(n)
		if err != nil {
			break
		}
		err = uc.sendSubscribed(ctx, n, cmd)
	case domain.NotificationFlightUpdated:
		var cmd command.SendFlightUpdatedCommand
		cmd, err = command.NewSendFlightUpdatedCommand(n)
		if err != nil {
			break
		}
		err = uc.sendFlightUpdated(ctx, n, cmd)
	default:
		err = fmt.Errorf("unsupported notification type %q", n.Type)
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "SMTP delivery failed")
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (uc *EmailSenderUsecase) sendSubscribed(
	ctx context.Context,
	n domain.Notification,
	cmd command.SendSubscribedNotificationCommand,
) error {
	var subj, body string
	subj = "Flight departure"
	body = formDepartureBody(cmd)

	msg := []byte(
		"From: " + uc.appEmail.String() + sep +
			"To: " + cmd.ToEmail.String() + sep +
			"Subject: " + subj + sep +
			sep +
			body + sep,
	)

	if err := uc.m.SendEmail(ctx, []common.Email{cmd.ToEmail}, msg, stableMessageID(n, cmd.ToEmail)); err != nil {
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
	n domain.Notification,
	cmd command.SendFlightUpdatedCommand,
) error {
	var subj, body string
	subj = "Flight departure"
	body = formFlightUpdatedBody(cmd)

	if len(cmd.Users) != 1 {
		return fmt.Errorf("expected one recipient per notification, got %d", len(cmd.Users))
	}
	to := cmd.Users[0]
	msg := []byte(
		"From: " + uc.appEmail.String() + sep +
			"To: " + to.String() + sep +
			"Subject: " + subj + sep +
			sep +
			body + sep,
	)
	if err := uc.m.SendEmail(ctx, []common.Email{to}, msg, stableMessageID(n, to)); err != nil {
		return err
	}
	return nil
}

func stableMessageID(n domain.Notification, recipient common.Email) string {
	hash := sha256.Sum256([]byte(recipient.String()))
	return fmt.Sprintf("<%s.%s@airline-tracker>", n.Id.String(), hex.EncodeToString(hash[:8]))
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

	if cmd.Plan != nil {
		fmt.Fprintf(&body, "New flight's plan is %s"+sep, *cmd.Plan)
	}

	return body.String()
}
