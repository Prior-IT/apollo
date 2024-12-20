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
	db := tests.DB(t)
	service := postgres.NewUserService(db)
	tests.DeleteAllUsers(service)
	defer tests.DeleteAllUsers(service)
	ctx := context.Background()

	t.Run("ok: create user", func(t *testing.T) {
		email1, err := core.ParseEmailAddress(tests.Faker.Email())
		tests.Check(err)
		var email2 *core.EmailAddress
		// Make sure both e-mail addresses are unique
		for email2 == nil || email1.String() == email2.String() {
			email2, err = core.ParseEmailAddress(tests.Faker.Email())
			tests.Check(err)
		}
		for _, data := range []CreateUserTest{
			{
				name:        tests.Faker.Name(),
				email:       *email1,
				description: "name and email set",
			},
			{
				name:        "",
				email:       *email2,
				description: "name empty",
			},
		} {
			user, err := service.CreateUser(ctx, data.name, data.email, "nl")
			tests.Check(err)
			assert.NotNil(t, user, data.description)
			assert.Equal(t, data.name, user.Name, data.description)
			assert.Equal(t, data.email, user.Email, data.description)

		}
	})

	t.Run("err: create duplicate user", func(t *testing.T) {
		email, err := core.ParseEmailAddress(tests.Faker.Email())
		tests.Check(err)
		user1, err := service.CreateUser(ctx, tests.Faker.Name(), *email, "nl")
		tests.Check(err)
		assert.NotNil(t, user1, "The first user should be created correctly")

		user2, err := service.CreateUser(ctx, tests.Faker.Name(), *email, "nl")
		assert.NotNil(t, err, "The duplicate user should return an error")
		assert.Nil(t, user2, "The duplicate user should return nil for the user")
		assert.ErrorIs(t, err, core.ErrConflict, "The duplicate user should return ErrConflict")
	})

	t.Run("ok: get user", func(t *testing.T) {
		email, err := core.ParseEmailAddress("getuserok@example.com")
		tests.Check(err)
		user, err := service.CreateUser(ctx, tests.Faker.Name(), *email, "nl")
		tests.Check(err)

		user2, err := service.GetUser(ctx, user.ID)
		tests.Check(err)

		assert.Equal(t, user, user2)
	})

	t.Run("ok: delete user", func(t *testing.T) {
		email, err := core.ParseEmailAddress("deleteuserok@example.com")
		tests.Check(err)
		user, err := service.CreateUser(ctx, tests.Faker.Name(), *email, "nl")
		tests.Check(err)

		tests.Check(service.DeleteUser(ctx, user.ID))

		user2, err := service.GetUser(ctx, user.ID)
		assert.NotNil(t, err, "Getting a deleted user should return an error")
		assert.Nil(t, user2, "Getting a deleted user should return nil for the user")
		assert.ErrorIs(t, err, core.ErrNotFound, "Getting a deleted user should return ErrNotFound")
	})

	t.Run("ok: update user admin", func(t *testing.T) {
		email, err := core.ParseEmailAddress("updateadminok@example.com")
		tests.Check(err)
		user, err := service.CreateUser(ctx, tests.Faker.Name(), *email, "nl")
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

	t.Run("ok: update user - empty", func(t *testing.T) {
		email, err := core.ParseEmailAddress("updateuseremptyok@example.com")
		tests.Check(err)
		name := tests.Faker.Name()
		user, err := service.CreateUser(ctx, name, *email, "nl")
		tests.Check(err)
		assert.Equal(t, name, user.Name)
		assert.Equal(t, *email, user.Email)

		user, err = service.UpdateUser(ctx, user.ID, core.UserUpdate{})

		tests.Check(err)
		assert.NotNil(t, user, "User should still exist after update")
		assert.Equal(t, name, user.Name, "Name should not change after empty update")
		assert.Equal(t, *email, user.Email, "Email should not change after empty update")
	})

	t.Run("ok: update user - name", func(t *testing.T) {
		email, err := core.ParseEmailAddress("updateusernameok@example.com")
		tests.Check(err)
		name := tests.Faker.Name()
		user, err := service.CreateUser(ctx, name, *email, "nl")
		tests.Check(err)
		assert.Equal(t, name, user.Name)
		assert.Equal(t, *email, user.Email)

		newName := tests.Faker.Name()
		user, err = service.UpdateUser(ctx, user.ID, core.UserUpdate{Name: &newName})

		tests.Check(err)
		assert.NotNil(t, user, "User should still exist after update")
		assert.Equal(t, newName, user.Name, "Name should have changed after name update")
		assert.Equal(t, *email, user.Email, "Email should not change after name update")
	})

	t.Run("ok: update user - email", func(t *testing.T) {
		email, err := core.ParseEmailAddress("updateuseremailok@example.com")
		tests.Check(err)
		name := tests.Faker.Name()
		user, err := service.CreateUser(ctx, name, *email, "nl")
		tests.Check(err)
		assert.Equal(t, name, user.Name)
		assert.Equal(t, *email, user.Email)

		// Email update
		newEmail, err := core.ParseEmailAddress("updateuseremailok2@example.com")
		tests.Check(err)
		newEmailStr := newEmail.String()
		user, err = service.UpdateUser(ctx, user.ID, core.UserUpdate{Email: &newEmailStr})

		tests.Check(err)
		assert.NotNil(t, user, "User should still exist after update")
		assert.Equal(t, name, user.Name, "Name should not change after email update")
		assert.Equal(t, *newEmail, user.Email, "Email should change after email update")
	})

	t.Run("ok: update user - name + email", func(t *testing.T) {
		email, err := core.ParseEmailAddress("updateuserbothok@example.com")
		tests.Check(err)
		name := tests.Faker.Name()
		user, err := service.CreateUser(ctx, name, *email, "nl")
		tests.Check(err)
		assert.Equal(t, name, user.Name)
		assert.Equal(t, *email, user.Email)

		newName := tests.Faker.Name()
		newEmail, err := core.ParseEmailAddress("updateuserbothok2@example.com")
		tests.Check(err)
		newEmailStr := newEmail.String()
		user, err = service.UpdateUser(
			ctx,
			user.ID,
			core.UserUpdate{Name: &newName, Email: &newEmailStr},
		)

		tests.Check(err)
		assert.NotNil(t, user, "User should still exist after update")
		assert.Equal(t, newName, user.Name, "Name should change after update")
		assert.Equal(t, *newEmail, user.Email, "Email should change after update")
	})

	t.Run("ok: update user - language", func(t *testing.T) {
		email, err := core.ParseEmailAddress("updateuseremailok@example.com")
		assert.Nil(t, err)
		lang := tests.Faker.Language()
		user, err := service.CreateUser(ctx, tests.Faker.Name(), *email, lang)
		assert.Nil(t, err)
		assert.Equal(t, lang, user.Lang)
		assert.Equal(t, *email, user.Email)

		// Lang update
		newLang := tests.Faker.Language()
		assert.Nil(t, err)
		user, err = service.UpdateUser(ctx, user.ID, core.UserUpdate{Lang: &newLang})
		assert.Nil(t, err)

		assert.NotNil(t, user, "User should still exist after update")
		assert.Equal(t, *email, user.Email, "Email should not change after lang update")
		assert.Equal(t, newLang, user.Lang, "Language should change after lang update")
	})
}
