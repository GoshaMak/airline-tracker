package pkg

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"shared/common"
	"strings"
	"time"
)

const mailingTimeout = 10 * time.Second

type Mailer interface {
	SendEmail(ctx context.Context, recipient common.Email, subject, body string) error
}

type mailingClient struct {
	address string
	host    string
	from    common.Email
	auth    smtp.Auth
	dialer  *net.Dialer
	timeout time.Duration
}

func NewMailingClient(address, appEmail, username, password string) (Mailer, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return nil, errors.New("smtp addr must be configured")
	}
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("invalid smtp addr: %w", err)
	}
	from, err := common.NewEmail(appEmail)
	if err != nil {
		return nil, fmt.Errorf("invalid app email: %w", err)
	}

	var auth smtp.Auth
	if username != "" {
		if password == "" {
			return nil, errors.New("smtp password must be configured when smtp username is set")
		}
		auth = smtp.PlainAuth("", username, password, host)
	}

	return &mailingClient{
		address: address,
		host:    host,
		from:    from,
		auth:    auth,
		dialer: &net.Dialer{
			Timeout: 5 * time.Second,
		},
		timeout: mailingTimeout,
	}, nil
}

func (m *mailingClient) SendEmail(ctx context.Context, recipient common.Email, subject, body string) error {
	message, err := newMessageBuilder().
		WithFrom(m.from).
		WithTo(recipient).
		WithSubject(subject).
		WithMIMEVersion("1.0").
		WithContentType("text/plain; charset=UTF-8").
		WithBody(body).
		Build()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()
	conn, err := m.dialer.DialContext(ctx, "tcp", m.address)
	if err != nil {
		return fmt.Errorf("SMTP dial: %w", err)
	}
	defer conn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return fmt.Errorf("SMTP deadline: %w", err)
		}
	}
	stop := context.AfterFunc(ctx, func() {
		_ = conn.SetDeadline(time.Now())
	})
	defer stop()

	client, err := smtp.NewClient(conn, m.host)
	if err != nil {
		return fmt.Errorf("SMTP handshake: %w", err)
	}
	defer client.Close()
	if supportsTLS, _ := client.Extension("STARTTLS"); supportsTLS {
		tlsCfg := &tls.Config{
			ServerName: m.host,
			MinVersion: tls.VersionTLS12,
		}
		if err := client.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("SMTP STARTTLS: %w", err)
		}
	}
	if m.auth != nil {
		if supportsAuth, _ := client.Extension("AUTH"); !supportsAuth {
			return errors.New("SMTP auth configured but server does not advertise AUTH")
		}
		if err := client.Auth(m.auth); err != nil {
			return fmt.Errorf("SMTP auth: %w", err)
		}
	}
	if err := client.Mail(m.from.String()); err != nil {
		return fmt.Errorf("SMTP sender: %w", err)
	}
	if err := client.Rcpt(recipient.String()); err != nil {
		return fmt.Errorf("SMTP recipient: %w", err)
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP data: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return fmt.Errorf("SMTP write: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("SMTP close data: %w", err)
	}
	return nil
}

type messageBuilder struct {
	from        common.Email
	to          common.Email
	subject     string
	mimeVersion string
	contentType string
	body        string
}

func newMessageBuilder() *messageBuilder {
	return &messageBuilder{}
}

func (b *messageBuilder) WithFrom(from common.Email) *messageBuilder {
	b.from = from
	return b
}

func (b *messageBuilder) WithTo(to common.Email) *messageBuilder {
	b.to = to
	return b
}

func (b *messageBuilder) WithSubject(subject string) *messageBuilder {
	b.subject = subject
	return b
}

func (b *messageBuilder) WithMIMEVersion(version string) *messageBuilder {
	b.mimeVersion = version
	return b
}

func (b *messageBuilder) WithContentType(contentType string) *messageBuilder {
	b.contentType = contentType
	return b
}

func (b *messageBuilder) WithBody(body string) *messageBuilder {
	b.body = body
	return b
}

func (b *messageBuilder) Build() ([]byte, error) {
	if b.from.String() == "" || b.to.String() == "" {
		return nil, errors.New("mail from and to are required")
	}
	if strings.ContainsAny(b.subject, "\r\n") {
		return nil, errors.New("mail subject must not contain line breaks")
	}
	if b.mimeVersion == "" {
		return nil, errors.New("MIME version is required")
	}
	if b.contentType == "" {
		return nil, errors.New("content type is required")
	}

	var message strings.Builder
	fmt.Fprintf(&message, "From: %s\r\n", b.from.String())
	fmt.Fprintf(&message, "To: %s\r\n", b.to.String())
	fmt.Fprintf(&message, "Subject: %s\r\n", b.subject)
	fmt.Fprintf(&message, "MIME-Version: %s\r\n", b.mimeVersion)
	fmt.Fprintf(&message, "Content-Type: %s\r\n\r\n", b.contentType)
	message.WriteString(b.body)
	message.WriteString("\r\n")
	return []byte(message.String()), nil
}
