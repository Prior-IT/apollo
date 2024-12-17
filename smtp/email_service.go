package smtp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strings"

	"github.com/a-h/templ"
	"github.com/prior-it/apollo/config"
	"github.com/prior-it/apollo/core"
	"gopkg.in/gomail.v2"
)

type EmailService struct {
	from          string
	d             *gomail.Dialer
	notifications *core.EmailAddress
}

func NewEmailService(cfg config.EmailConfig) (*EmailService, error) {
	d := gomail.NewDialer(
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
	)

	s := EmailService{
		from: cfg.From,
		d:    d,
	}

	address, err := core.ParseEmailAddress(cfg.Notifications)
	switch {
	case errors.Is(err, core.ErrEmailAddressEmpty):
		slog.Info("Notification e-mail address empty, notifications will not be sent")
	case err != nil:
		return nil, err
	default:
		s.notifications = address
	}

	return &s, nil
}

// SendEmail will build and send a basic e-mail message containing both a HTML template version (optional) as well as a plaintext alternative (required).
// This will open a new server connection and immediately close it after sending the e-mail.
func (s *EmailService) SendEmail(
	ctx context.Context,
	address core.EmailAddress,
	subject string,
	template *templ.Component,
	plaintextMessage string,
) error {
	m := gomail.NewMessage()

	m.SetHeader("From", s.from)
	m.SetHeader("To", address.String())
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", plaintextMessage)

	if template != nil {
		var builder strings.Builder
		err := (*template).Render(ctx, &builder)
		if err != nil {
			log.Fatal(err)
		}
		m.AddAlternative("text/html", builder.String())
	}

	return s.d.DialAndSend(m)
}

// SendNotification will send a specific plain-text notification to the configured notification
// address.
func (s *EmailService) SendNotification(
	ctx context.Context,
	subject string,
	message string,
	args ...any,
) error {
	if s.notifications == nil {
		return core.ErrEmailAddressEmpty
	}
	return s.SendEmail(
		ctx,
		*s.notifications,
		subject,
		nil,
		fmt.Sprintf(message, args...),
	)
}

// SendRawMessage will send a raw gomail message using the existing smtp connection.
// Note that this overrides the "From" header to use the configured value.
// This will open a new server connection and immediately close it after sending the e-mail.
func (s *EmailService) SendRawMessage(
	_ context.Context,
	message *gomail.Message,
) error {
	message.SetHeader("From", s.from)
	return s.d.DialAndSend(message)
}
