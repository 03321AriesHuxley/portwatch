package alert

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

// EmailConfig holds the SMTP configuration for sending email alerts.
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       []string
}

// emailNotifier sends alert events via SMTP email.
type emailNotifier struct {
	cfg  EmailConfig
	auth smtp.Auth
	send func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

// EmailOption is a functional option for emailNotifier.
type EmailOption func(*emailNotifier)

// WithEmailSendFunc overrides the smtp.SendMail function (used in tests).
func WithEmailSendFunc(fn func(string, smtp.Auth, string, []string, []byte) error) EmailOption {
	return func(e *emailNotifier) {
		e.send = fn
	}
}

// NewEmailNotifier creates a Notifier that sends events via email.
func NewEmailNotifier(cfg EmailConfig, opts ...EmailOption) Notifier {
	n := &emailNotifier{
		cfg:  cfg,
		auth: smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host),
		send: smtp.SendMail,
	}
	for _, o := range opts {
		o(n)
	}
	return n
}

func (e *emailNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	subject := fmt.Sprintf("portwatch: %d port change(s) detected", len(events))
	body := FormatEvents(events)

	msg := []byte(strings.Join([]string{
		"From: " + e.cfg.From,
		"To: " + strings.Join(e.cfg.To, ", "),
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=utf-8",
		"",
		body,
	}, "\r\n"))

	addr := fmt.Sprintf("%s:%d", e.cfg.Host, e.cfg.Port)
	return e.send(addr, e.auth, e.cfg.From, e.cfg.To, msg)
}
