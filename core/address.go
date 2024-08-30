package core

import (
	"context"
	"errors"
	"strconv"
)

/**
 * DOMAIN
 */

type Address struct {
	ID         AddressID
	Street     string
	Number     string
	PostalCode string
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

/**
 * APPLICATION
 */

type AddressUpdateData struct {
	Street     *string
	Number     *string
	PostalCode *string
	City       *string
	Country    *string
	ExtraLine  *string
}
type AddressService interface {
	// Creates an Address from an Address struct, the ID field of the struct gets ignored here
	CreateAddress(ctx context.Context, address Address) (*Address, error)
	GetAddress(ctx context.Context, addressID AddressID) (*Address, error)
	DeleteAddress(ctx context.Context, addressID AddressID) error
	UpdateAddress(ctx context.Context, addressID AddressID, update AddressUpdateData) (*Address, error)
	ListAddresses(ctx context.Context) ([]Address, error)
}
