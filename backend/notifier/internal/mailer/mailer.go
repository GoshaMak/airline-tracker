package mailer

import (
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"os"
	"shared/common"

	"github.com/samber/do/v2"
)

type Mailer struct {
	Addr     string
	AppEmail common.Email
	Auth     smtp.Auth
}

func NewMailer(i do.Injector) (*Mailer, error) {
	const op = "NewMailer"
	addr := os.Getenv("SMTP_ADDR")
	appEmail, err := common.NewEmail(os.Getenv("APP_EMAIL"))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var auth smtp.Auth
	if addr != "" {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("%s: invalid SMTP_ADDR: %w", op, err)
		}
		if password := os.Getenv("APP_PASSWORD"); password != "" {
			auth = smtp.PlainAuth("", appEmail.String(), password, host)
		}
	}

	return &Mailer{
		Addr:     addr,
		AppEmail: appEmail,
		Auth:     auth,
	}, nil
}

func (m *Mailer) SendEmail(to []common.Email, msg []byte) error {
	const op = "Emailer.SendEmail"

	toEmails := []string{}
	for _, email := range to {
		toEmails = append(toEmails, email.String())
	}
	slog.Debug(op, "targets", toEmails, "message", string(msg))

	if m.Addr == "" {
		slog.Info("SMTP delivery disabled: SMTP_ADDR is not configured")
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
