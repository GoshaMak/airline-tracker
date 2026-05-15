package usecase

import (
	"fmt"
	"notifier/internal/mailer"
	"notifier/internal/notifier/command"
	"shared/common"
	"time"
)

type NotifierUsecase struct {
	m *mailer.Mailer
}

func NewNotifierUsecase(from common.Email, appPassword string) (*NotifierUsecase, error) {
	const op = "NewNotifierUsecase"
	m, err := mailer.NewMailer(from, appPassword)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &NotifierUsecase{
		m: m,
	}, nil
}

func (u *NotifierUsecase) SendNotification(cmd command.SendNotificationCommand) error {
	const op = "NotifierUsecase.SendNotification"
	const sep = "\r\n"
	toEmail, err := common.NewEmail(cmd.To)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	subj := "Flight departure"
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

	msg := []byte(
		"From: " + u.m.From.String() + sep +
			"To: " + toEmail.String() + sep +
			"Subject: " + subj + sep +
			sep +
			body + sep,
	)

	if err := u.m.SendEmail([]common.Email{toEmail}, msg); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
