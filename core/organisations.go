package core

import (
	"errors"
	"fmt"
	"strconv"
)

/**
 * DOMAIN
 */

type Organisation struct {
	ID      OrganisationID
	Name    string
}

type (
	OrganisationID uint
)

func (id OrganisationID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

// NewOrganisationID parses an organisation id from any unsigned integer.
func NewOrganisationID(id uint) (OrganisationID, error) {
	if id == 0 {
		return 0, errors.New("OrganisationID cannot be 0")
	}
	return OrganisationID(id), nil
}

// ParseOrganisationID parses a string into an organisation id.
func ParseOrganisationID(id string) (OrganisationID, error) {
	integerID, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("cannot parse organisation id: %w", err)
	}
	if integerID < 0 {
		return 0, errors.New("cannot parse organisation id: organisation ids cannot be negative")
	}
	organisationID, err := NewOrganisationID(uint(integerID))
	if err != nil {
		return 0, fmt.Errorf("cannot parse organisation id: %w", err)
	}
	return organisationID, nil
}
