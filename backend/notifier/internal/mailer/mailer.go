package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"os"
	"shared/common"
	"time"

	"github.com/samber/do/v2"
)

type Mailer struct {
	Addr     string
	AppEmail common.Email
	Auth     smtp.Auth
	dialer   contextDialer
	timeout  time.Duration
}

type contextDialer interface {
	DialContext(context.Context, string, string) (net.Conn, error)
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
		dialer:   &net.Dialer{Timeout: 5 * time.Second},
		timeout:  5 * time.Second,
	}, nil
}

func (m *Mailer) SendEmail(ctx context.Context, to []common.Email, msg []byte, messageID string) error {
	const op = "Emailer.SendEmail"

	toEmails := []string{}
	for _, email := range to {
		toEmails = append(toEmails, email.String())
	}
	if m.Addr == "" {
		return nil
	}
	if len(toEmails) == 0 {
		return fmt.Errorf("%s: no recipient", op)
	}
	if m.timeout <= 0 {
		m.timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()
	dialer := m.dialer
	if dialer == nil {
		dialer = &net.Dialer{Timeout: m.timeout}
	}
	conn, err := dialer.DialContext(ctx, "tcp", m.Addr)
	if err != nil {
		return fmt.Errorf("%s dial: %w", op, err)
	}
	defer conn.Close()
	deadline := time.Now().Add(m.timeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = conn.SetDeadline(deadline)
	stop := context.AfterFunc(ctx, func() { _ = conn.SetDeadline(time.Now()) })
	defer stop()

	host, _, err := net.SplitHostPort(m.Addr)
	if err != nil {
		return fmt.Errorf("%s invalid SMTP address: %w", op, err)
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("%s handshake: %w", op, err)
	}
	defer client.Close()
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return fmt.Errorf("%s STARTTLS: %w", op, err)
		}
	}
	if m.Auth != nil {
		if ok, _ := client.Extension("AUTH"); !ok {
			return fmt.Errorf("%s auth configured but SMTP server does not advertise AUTH", op)
		}
		if err := client.Auth(m.Auth); err != nil {
			return fmt.Errorf("%s auth: %w", op, err)
		}
	}
	if err := client.Mail(m.AppEmail.String()); err != nil {
		return fmt.Errorf("%s sender: %w", op, err)
	}
	for _, recipient := range toEmails {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("%s recipient: %w", op, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("%s data: %w", op, err)
	}
	if _, err := w.Write(append([]byte("Message-ID: "+messageID+"\r\n"), msg...)); err != nil {
		_ = w.Close()
		return fmt.Errorf("%s write: %w", op, err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("%s close data: %w", op, err)
	}
	if err := client.Quit(); err != nil {
		slog.WarnContext(ctx, "SMTP QUIT failed after message acceptance", "err", err)
	}
	return nil
}
