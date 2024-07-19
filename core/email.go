package core

import (
	"fmt"
	"net/mail"
)

type EmailAddress interface {
	// ToString returns the string representation for this e-mail address
	ToString() string
}

type emailAddress struct {
	address string
}

func (email emailAddress) ToString() string {
	return email.address
}

// NewEmailAddress parses an e-mail address from any string.
// This uses RFC-5322 to determine valid e-mail addresses, e.g. "Biggie Smalls <notorious@example.com>"
func NewEmailAddress(address string) (EmailAddress, error) {
	_, err := mail.ParseAddress(address)
	if err != nil {
		return nil, fmt.Errorf("cannot parse e-mail address %q: %w", address, err)
	}
	return emailAddress{address}, nil
}
