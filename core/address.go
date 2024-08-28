package core

import (
	"errors"
	"fmt"
	"strconv"
)

/**
 * DOMAIN
 */

type Address struct {
	ID         AddressID
	Street     string
	Number     uint
	PostalCode uint
	City       string
	Country    string
	ExtraLine  *string
}

type (
	AddressID uint
)

func (id AddressID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

// NewAddressID parses an address id from any unsigned integer.
func NewAddressID(id uint) (AddressID, error) {
	if id == 0 {
		return 0, errors.New("AddressID cannot be 0")
	}
	return AddressID(id), nil
}

// ParseAddressID parses a string into an address id.
func ParseAddressID(id string) (AddressID, error) {
	integerID, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("cannot parse address id: %w", err)
	}
	if integerID < 0 {
		return 0, errors.New("cannot parse address id: address ids cannot be negative")
	}
	addressID, err := NewAddressID(uint(integerID))
	if err != nil {
		return 0, fmt.Errorf("cannot parse address id: %w", err)
	}
	return addressID, nil
}
