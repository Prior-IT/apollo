package postgres

import (
	"context"

	"github.com/prior-it/apollo/core"
	"github.com/prior-it/apollo/postgres/internal/sqlc"
)

func NewAddressService(DB *ApolloDB) *AddressService {
	q := sqlc.New(DB)
	return &AddressService{DB, q}
}

// Postgres implementation of the core AddressService interface.
type AddressService struct {
	db *ApolloDB
	q  *sqlc.Queries
}

// Force struct to implement the core interface
var _ core.AddressService = &AddressService{}

// CreateAddress implements core.AddressService.CreateAddress
func (a *AddressService) CreateAddress(
	ctx context.Context,
	addressCreate core.AddressCreateData,
) (*core.Address, error) {
	return a.CreateAddressTx(ctx, a.db, addressCreate)
}

// CreateAddress implements core.AddressService.CreateAddress
func (a *AddressService) CreateAddressTx(
	ctx context.Context,
	dbtx sqlc.DBTX,
	addressCreate core.AddressCreateData,
) (*core.Address, error) {
	q := sqlc.New(dbtx)
	address, err := q.CreateAddress(ctx, sqlc.CreateAddressParams{
		Street:     addressCreate.Street,
		Number:     int32(addressCreate.Number),
		PostalCode: int32(addressCreate.PostalCode),
		City:       addressCreate.City,
		Country:    addressCreate.Country,
		ExtraLine:  addressCreate.ExtraLine,
	})
	if err != nil {
		return nil, convertPgError(err)
	}
	return convertAddress(address)
}

// DeleteAddress implements core.AddressService.DeleteAddress
func (a *AddressService) DeleteAddress(ctx context.Context, id core.AddressID) error {
	return a.q.DeleteAddress(ctx, int32(id))
}

// GetAddress implements core.AddressService.GetAddress
func (a *AddressService) GetAddress(ctx context.Context, id core.AddressID) (*core.Address, error) {
	address, err := a.q.GetAddress(ctx, int32(id))
	if err != nil {
		return nil, convertPgError(err)
	}
	return convertAddress(address)
}

// UpdateAddress implements core.AddressService.UpdateAddress
func (a *AddressService) UpdateAddress(
	ctx context.Context,
	id core.AddressID,
	data core.AddressUpdateData,
) (*core.Address, error) {
	dbAddress, err := a.q.UpdateAddress(ctx, sqlc.UpdateAddressParams{
		ID:         int32(id),
		Street:     data.Street,
		Number:     data.Number,
		PostalCode: data.PostalCode,
		City:       data.City,
		Country:    data.Country,
		ExtraLine:  data.ExtraLine,
	})
	if err != nil {
		return nil, err
	}
	return convertAddress(dbAddress)
}

func convertAddress(address sqlc.ApolloAddress) (*core.Address, error) {
	id, err := core.NewAddressID(uint(address.ID))
	if err != nil {
		return nil, err
	}
	return &core.Address{
		ID:         id,
		Street:     address.Street,
		Number:     uint(address.Number),
		PostalCode: uint(address.PostalCode),
		City:       address.City,
		Country:    address.Country,
		ExtraLine:  address.ExtraLine,
	}, nil
}
