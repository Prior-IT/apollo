package core

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/a-h/templ"
	"gopkg.in/gomail.v2"
)

var (
	ErrInvalidEmailAddress = errors.New("invalid e-mail address")
	ErrEmailAddressEmpty   = errors.New("e-mail address is empty")
)

type EmailAddress interface {
	// String returns the string representation for this e-mail address
	String() string
}

type emailAddress struct {
	address string
}

func (email emailAddress) String() string {
	return email.address
}

// NewEmailAddress parses an e-mail address from any string.
// This uses RFC-5322 to determine valid e-mail addresses, e.g. "Biggie Smalls <notorious@example.com>"
func NewEmailAddress(address string) (EmailAddress, error) {
	if len(address) == 0 {
		return nil, errors.Join(ErrInvalidEmailAddress, ErrEmailAddressEmpty)
	}
	address = strings.ToLower(address)
	_, err := mail.ParseAddress(address)
	if err != nil {
		return nil, errors.Join(
			ErrInvalidEmailAddress,
			fmt.Errorf("cannot parse e-mail address %q: %w", address, err),
		)
	}
	return emailAddress{address}, nil
}

type EmailService interface {
	// SendEmail will build and send a basic e-mail message containing both a HTML template version
	// (optional) as well as a plaintext alternative (required).
	SendEmail(
		ctx context.Context,
		address EmailAddress,
		subject string,
		template *templ.Component,
		plaintextMessage string,
	) error

	// SendNotification will send a specific plain-text notification to the configured notification
	// address.
	SendNotification(
		ctx context.Context,
		subject string,
		message string,
	) error

	// SendRawMessage will send a raw gomail message using the existing configuration.
	SendRawMessage(ctx context.Context, message *gomail.Message) error
}
