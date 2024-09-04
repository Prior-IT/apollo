package core

import (
	"context"
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

type AddressID = ID

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
