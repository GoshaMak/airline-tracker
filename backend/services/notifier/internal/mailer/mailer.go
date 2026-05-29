package mailer

import (
	"errors"
	"fmt"
	"log/slog"
	"net/smtp"
	"os"
	"shared/common"

	"github.com/samber/do/v2"
)

type Mailer struct {
	Addr        string
	AppEmail    common.Email
	AppPassword string
	Auth        smtp.Auth
}

func NewMailer(i do.Injector) (*Mailer, error) {
	const op = "NewMailer"
	appPassword := os.Getenv("APP_PASSWORD")
	if len(appPassword) == 0 {
		return nil, errors.New("empty app password")
	}
	appEmail, err := common.NewEmail(os.Getenv("APP_EMAIL"))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	auth := smtp.PlainAuth("", appEmail.String(), appPassword, "smtp.gmail.com")

	return &Mailer{
		Addr:        "smtp.gmail.com:587",
		AppEmail:    appEmail,
		AppPassword: appPassword,
		Auth:        auth,
	}, nil
}

func (m *Mailer) SendEmail(to []common.Email, msg []byte) error {
	const op = "Emailer.SendEmail"

	toEmails := []string{}
	for _, email := range to {
		toEmails = append(toEmails, email.String())
	}
	slog.Debug(op, "toEmails", toEmails, "msg", string(msg))

	slog.Info("sending email faked")
	return nil

	targetEmail := "je6t8r@gmail.com"
	if toEmails[0] != targetEmail {
		slog.Info("sending email faked")
		return nil
	}

	err := smtp.SendMail(
		m.Addr,
		m.Auth,
		m.AppEmail.String(),
		toEmails,
		msg,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
