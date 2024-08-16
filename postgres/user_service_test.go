package postgres_test

import (
	"context"
	"testing"

	"github.com/prior-it/apollo/core"
	"github.com/prior-it/apollo/postgres"
	"github.com/prior-it/apollo/tests"
	"github.com/stretchr/testify/assert"
)

type CreateUserTest struct {
	name        string
	email       core.EmailAddress
	description string
}

func TestUserService(t *testing.T) {
	db := tests.DB()
	service := postgres.NewUserService(db)
	defer tests.DeleteAllUsers(service)
	ctx := context.Background()

	t.Run("ok: create user", func(t *testing.T) {
		email1, err := core.NewEmailAddress(tests.Faker.Email())
		tests.Check(err)
		var email2 core.EmailAddress
		// Make sure both e-mail addresses are unique
		for email2 == nil || email1.String() == email2.String() {
			email2, err = core.NewEmailAddress(tests.Faker.Email())
			tests.Check(err)
		}
		for _, data := range []CreateUserTest{
			{
				name:        tests.Faker.Name(),
				email:       email1,
				description: "name and email set",
			},
			{
				name:        "",
				email:       email2,
				description: "name empty",
			},
			{
				name:        tests.Faker.Name(),
				email:       nil,
				description: "email empty",
			},
			{
				name:        "",
				email:       nil,
				description: "name and email empty",
			},
		} {
			user, err := service.CreateUser(ctx, core.UserCreateData{
				Name:  data.name,
				Email: data.email,
			})
			if data.email != nil {
				tests.Check(err)
				assert.NotNil(t, user, data.description)
				assert.Equal(t, data.name, user.Name, data.description)
				assert.Equal(t, data.email, user.Email, data.description)
			} else {
				assert.Error(t, err, "Nil email should return an error")
			}
		}
	})

	t.Run("err: create duplicate user", func(t *testing.T) {
		email, err := core.NewEmailAddress(tests.Faker.Email())
		tests.Check(err)
		user1, err := service.CreateUser(ctx, core.UserCreateData{
			Name:  tests.Faker.Name(),
			Email: email,
		})
		tests.Check(err)
		assert.NotNil(t, user1, "The first user should be created correctly")

		user2, err := service.CreateUser(ctx, core.UserCreateData{
			Name:  tests.Faker.Name(),
			Email: email,
		})
		assert.NotNil(t, err, "The duplicate user should return an error")
		assert.Nil(t, user2, "The duplicate user should return nil for the user")
		assert.ErrorIs(t, err, core.ErrConflict, "The duplicate user should return ErrConflict")
	})

	t.Run("ok: get user", func(t *testing.T) {
		email, err := core.NewEmailAddress("getuserok@example.com")
		tests.Check(err)
		user, err := service.CreateUser(ctx, core.UserCreateData{
			Name:  tests.Faker.Name(),
			Email: email,
		})
		tests.Check(err)

		user2, err := service.GetUser(ctx, user.ID)
		tests.Check(err)

		assert.Equal(t, user, user2)
	})

	t.Run("ok: delete user", func(t *testing.T) {
		email, err := core.NewEmailAddress("deleteuserok@example.com")
		tests.Check(err)
		user, err := service.CreateUser(ctx, core.UserCreateData{
			Name:  tests.Faker.Name(),
			Email: email,
		})
		tests.Check(err)

		tests.Check(service.DeleteUser(ctx, user.ID))

		user2, err := service.GetUser(ctx, user.ID)
		assert.NotNil(t, err, "Getting a deleted user should return an error")
		assert.Nil(t, user2, "Getting a deleted user should return nil for the user")
		assert.ErrorIs(t, err, core.ErrNotFound, "Getting a deleted user should return ErrNotFound")
	})

	t.Run("ok: update user admin", func(t *testing.T) {
		email, err := core.NewEmailAddress("updateadminok@example.com")
		tests.Check(err)
		user, err := service.CreateUser(ctx, core.UserCreateData{
			Name:  tests.Faker.Name(),
			Email: email,
		})
		tests.Check(err)
		assert.False(t, user.Admin, "Admin should be false by default")

		tests.Check(service.UpdateUserAdmin(ctx, user.ID, true))

		user2, err := service.GetUser(ctx, user.ID)
		tests.Check(err)
		assert.NotNil(t, user2, "User should still exist after update")
		assert.True(t, user2.Admin, "Admin should be true after update")

		tests.Check(service.UpdateUserAdmin(ctx, user.ID, false))

		user2, err = service.GetUser(ctx, user.ID)
		tests.Check(err)
		assert.NotNil(t, user2, "User should still exist after update")
		assert.False(t, user2.Admin, "Admin should be back to false after second update")
	})
}
