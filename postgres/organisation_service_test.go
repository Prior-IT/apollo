package postgres_test

import (
	"context"
	"testing"

	"github.com/prior-it/apollo/core"
	"github.com/prior-it/apollo/postgres"
	"github.com/prior-it/apollo/tests"
	"github.com/stretchr/testify/assert"
)

func TestOrganisationService(t *testing.T) {
	db := tests.DB(t)
	service := postgres.NewOrganisationService(db)
	UserService := postgres.NewUserService(db)
	defer tests.DeleteAllOrganisations(service)
	ctx := context.Background()

	t.Run("ok: create duplicate organisation", func(t *testing.T) {
		name := tests.Faker.BS()
		organisation1, err := service.CreateOrganisation(ctx, name, nil)
		tests.Check(err)
		assert.NotNil(t, organisation1, "The first organisation should be created correctly")

		organisation2, err := service.CreateOrganisation(ctx, name, nil)
		tests.Check(err)
		assert.NotNil(t, organisation2, "The second organisation should be created correctly")
	})

	t.Run("ok: create multiple organisations with parent", func(t *testing.T) {
		organisation1, err := service.CreateOrganisation(ctx, tests.Faker.BS(), nil)
		tests.Check(err)
		assert.NotNil(t, organisation1, "The first organisation should be created correctly")

		for idx := range 3 {
			newOrg, err := service.CreateOrganisation(ctx, tests.Faker.BS(), &organisation1.ID)
			tests.Check(err)
			assert.NotNilf(t, newOrg, "Organisation with index %d should be created correctly", idx)
			assert.Equalf(
				t,
				*newOrg.ParentID,
				organisation1.ID,
				"Organisation 1 should be the parent of organisation with index %d",
				idx,
			)
		}
	})

	t.Run("ok: get organisation", func(t *testing.T) {
		name := tests.Faker.BS()
		organisation, err := service.CreateOrganisation(ctx, name, nil)
		tests.Check(err)

		organisation2, err := service.GetOrganisation(ctx, organisation.ID)
		tests.Check(err)

		assert.Equal(t, organisation, organisation2)
	})

	t.Run("ok: update organisation", func(t *testing.T) {
		name := tests.Faker.BS()
		organisation, err := service.CreateOrganisation(ctx, name, nil)
		tests.Check(err)

		organisation.Name = tests.Faker.BS()
		err = service.UpdateOrganisation(ctx, organisation.ID, organisation.Name)
		tests.Check(err)

		organisation2, err := service.GetOrganisation(ctx, organisation.ID)
		tests.Check(err)

		assert.Equal(t, organisation, organisation2)
	})

	t.Run("ok: delete organisation", func(t *testing.T) {
		name := tests.Faker.BS()
		organisation, err := service.CreateOrganisation(ctx, name, nil)
		tests.Check(err)

		tests.Check(service.DeleteOrganisation(ctx, organisation.ID))

		organisation2, err := service.GetOrganisation(ctx, organisation.ID)
		assert.NotNil(t, err, "Getting a deleted organisation should return an error")
		assert.Nil(
			t,
			organisation2,
			"Getting a deleted organisation should return nil for the organisation",
		)
		assert.ErrorIs(
			t,
			err,
			core.ErrNotFound,
			"Getting a deleted organisation should return ErrNotFound",
		)
	})

	t.Run("ok: delete organisation removes children", func(t *testing.T) {
		org1, err := service.CreateOrganisation(ctx, tests.Faker.BS(), nil)
		tests.Check(err)
		org2, err := service.CreateOrganisation(ctx, tests.Faker.BS(), &org1.ID)
		tests.Check(err)
		tests.Check(service.DeleteOrganisation(ctx, org1.ID))

		org2bis, err := service.GetOrganisation(ctx, org2.ID)
		assert.NotNil(t, err, "Getting an organisation with deleted parent should return an error")
		assert.Nil(
			t,
			org2bis,
			"Getting an organisation with deleted parent should return nil for the organisation",
		)
		assert.ErrorIs(
			t,
			err,
			core.ErrNotFound,
			"Getting an organisation with deleted parent should return ErrNotFound",
		)
	})

	t.Run("ok: list organisation children return only own children", func(t *testing.T) {
		org1, err := service.CreateOrganisation(ctx, tests.Faker.BS(), nil)
		tests.Check(err)
		org1A, err := service.CreateOrganisation(ctx, tests.Faker.BS(), &org1.ID)
		tests.Check(err)
		org1B, err := service.CreateOrganisation(ctx, tests.Faker.BS(), &org1.ID)
		tests.Check(err)
		org2, err := service.CreateOrganisation(ctx, tests.Faker.BS(), nil)
		tests.Check(err)
		_, err = service.CreateOrganisation(ctx, tests.Faker.BS(), &org2.ID)
		tests.Check(err)

		children, err := service.ListOrganisationChildren(ctx, org1.ID)
		childIDs := []core.OrganisationID{}
		for _, child := range children {
			childIDs = append(childIDs, child.ID)
		}
		tests.Check(err)
		assert.Equal(t, len(children), 2)
		assert.Contains(t, childIDs, org1A.ID)
		assert.Contains(t, childIDs, org1B.ID)
	})

	t.Run("ok: add user to organisation and list", func(t *testing.T) {
		name := tests.Faker.BS()
		organisation, err := service.CreateOrganisation(ctx, name, nil)
		tests.Check(err)

		// List without users
		users, err := service.ListUsersInOrganisation(ctx, organisation.ID)
		assert.Nil(t, err, "Getting users in organisation should not error")
		assert.Len(t, users, 0, "Users list shoud be empty")

		// Add user to organisation
		email, err := core.NewEmailAddress("getuserok@example.com")
		tests.Check(err)
		user, err := UserService.CreateUser(ctx, tests.Faker.Name(), email)
		tests.Check(err)

		tests.Check(service.AddUser(ctx, user.ID, organisation.ID))

		// List with user in organisation
		users, err = service.ListUsersInOrganisation(ctx, organisation.ID)
		assert.Nil(t, err, "Getting users in organisation should not error")
		assert.NotEmpty(t, users, "Users list shoud not be empty")
		assert.Equal(t, users[0], *user)

		// Remove user from organisation
		tests.Check(service.RemoveUser(ctx, user.ID, organisation.ID))
		users, err = service.ListUsersInOrganisation(ctx, organisation.ID)
		assert.Nil(t, err, "Getting users in organisation should not error")
		assert.Len(t, users, 0, "Users list shoud be empty")
	})

	t.Run("ok: deleting user removes user from organisation", func(t *testing.T) {
		name := tests.Faker.BS()
		organisation, err := service.CreateOrganisation(ctx, name, nil)
		tests.Check(err)

		// Add user to organisation
		email, err := core.NewEmailAddress("deleteuserok@example.com")
		tests.Check(err)
		user, err := UserService.CreateUser(ctx, tests.Faker.Name(), email)
		tests.Check(err)

		tests.Check(service.AddUser(ctx, user.ID, organisation.ID))

		// Delete user
		tests.Check(UserService.DeleteUser(ctx, user.ID))
		users, err := service.ListUsersInOrganisation(ctx, organisation.ID)
		assert.Nil(t, err, "Getting users in organisation should not error")
		assert.Len(t, users, 0, "Users list shoud be empty")
	})
}
