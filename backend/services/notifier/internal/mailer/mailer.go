package mailer

import (
	"errors"
	"fmt"
	"log/slog"
	"net/smtp"
	"shared/common"
)

type Mailer struct {
	Addr        string
	From        common.Email
	AppPassword string
	Auth        smtp.Auth
}

func NewMailer(from common.Email, appPassword string) (*Mailer, error) {
	// from := "trackerairline@gmail.com"

	if len(appPassword) == 0 {
		return nil, errors.New("empty appPassword")
	}

	auth := smtp.PlainAuth("", from.String(), appPassword, "smtp.gmail.com")

	return &Mailer{
		Addr:        "smtp.gmail.com:587",
		From:        from,
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

	targetEmail := "je6t8r@gmail.com"
	if toEmails[0] != targetEmail {
		slog.Info("sending email faked")
		return nil
	}

	err := smtp.SendMail(
		m.Addr,
		m.Auth,
		m.From.String(),
		toEmails,
		msg,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
