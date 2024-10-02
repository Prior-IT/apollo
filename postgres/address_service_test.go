package postgres_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/prior-it/apollo/core"
	"github.com/prior-it/apollo/postgres"
	"github.com/prior-it/apollo/tests"
	"github.com/stretchr/testify/assert"
)

func TestAddressService(t *testing.T) {
	db := tests.DB(t)
	service := postgres.NewAddressService(db)
	ctx := context.Background()
	defer tests.DeleteAllAddresses(service)

	t.Run("ok: create address", func(t *testing.T) {
		fakeAddress := tests.Faker.Address()
		createData := core.Address{
			Street:     fakeAddress.Street,
			Number:     tests.Faker.Zip(),
			PostalCode: fakeAddress.Zip,
			ExtraLine:  &fakeAddress.State,
			City:       fakeAddress.City,
			Country:    fakeAddress.Country,
		}
		address, err := service.CreateAddress(ctx, createData)
		tests.Check(err)
		assert.NotNil(t, address, "The first address should be created correctly")
		assert.Equal(
			t,
			*address.ExtraLine,
			fakeAddress.State,
			"The extraLine field should be created correctly",
		)

		createData.ExtraLine = nil
		address, err = service.CreateAddress(ctx, createData)
		tests.Check(err)
		assert.NotNil(t, address, "The second address should be created correctly")
		assert.Nil(t, address.ExtraLine, "ExtraLine field should be nil")
	})

	t.Run("ok: update address", func(t *testing.T) {
		fakeAddress := tests.Faker.Address()
		createData := core.Address{
			Street:     fakeAddress.Street,
			Number:     tests.Faker.Zip(),
			PostalCode: fakeAddress.Zip,
			City:       fakeAddress.City,
			Country:    fakeAddress.Country,
		}
		address, err := service.CreateAddress(ctx, createData)
		tests.Check(err)
		assert.NotNil(t, address, "The first address should be created correctly")

		newCity := "Ghent"
		updateData := core.AddressUpdateData{
			City: &newCity,
		}
		updatedAddress, err := service.UpdateAddress(ctx, address.ID, updateData)
		tests.Check(err)
		assert.NotNil(t, updatedAddress, "The address should still exist")
		assert.Equal(t, newCity, updatedAddress.City, "The City field should be updated correctly")

		// Check other fields are unchanged
		a := reflect.ValueOf(*address)
		u := reflect.ValueOf(*updatedAddress)
		typeOfAddress := a.Type()
		for i := 0; i < a.NumField(); i++ {
			if typeOfAddress.Field(i).Name == "City" {
				continue
			}
			assert.Equal(
				t,
				a.Field(i).Interface(),
				u.Field(i).Interface(),
				"Other fields should remain unchanged",
			)
		}
	})

	t.Run("ok: delete address", func(t *testing.T) {
		fakeAddress := tests.Faker.Address()
		createData := core.Address{
			Street:     fakeAddress.Street,
			Number:     tests.Faker.Zip(),
			PostalCode: fakeAddress.Zip,
			ExtraLine:  &fakeAddress.State,
			City:       fakeAddress.City,
			Country:    fakeAddress.Country,
		}
		address, err := service.CreateAddress(ctx, createData)
		tests.Check(err)

		tests.Check(service.DeleteAddress(ctx, address.ID))

		address, err = service.GetAddress(ctx, address.ID)
		assert.NotNil(t, err, "Getting a deleted address should return an error")
		assert.Nil(t, address, "Getting a deleted address should return nil for the address")
		assert.ErrorIs(
			t,
			err,
			core.ErrNotFound,
			"Getting a deleted address should return ErrNotFound",
		)
	})
}
