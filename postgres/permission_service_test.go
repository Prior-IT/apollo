package postgres_test

import (
	"context"
	"log"
	"testing"

	"github.com/prior-it/apollo/core"
	"github.com/prior-it/apollo/permissions"
	"github.com/prior-it/apollo/postgres"
	"github.com/prior-it/apollo/tests"
	"github.com/stretchr/testify/assert"
)

func CreateUserWithPermissions(
	t *testing.T,
	db *postgres.DB,
	Permissions map[permissions.Permission]bool,
) *core.User {
	ctx := context.Background()
	service := postgres.NewPermissionService(db)
	userService := postgres.NewUserService(db)

	group, err := service.CreatePermissionGroup(ctx, &permissions.PermissionGroup{
		Permissions: Permissions,
	})
	assert.Nil(t, err)

	user := tests.CreateRegularUser(userService)
	err = service.AddUserToPermissionGroup(ctx, user.ID, group.ID)
	assert.Nil(t, err)

	return user
}

func TestPermissionService(t *testing.T) {
	db := tests.DB(t)
	service := postgres.NewPermissionService(db)
	userService := postgres.NewUserService(db)
	defer tests.DeleteAllPermissions(service)
	defer tests.DeleteAllUsers(userService)

	err := permissions.RegisterApolloPermissions(service)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	t.Run("ok: user without group", func(t *testing.T) {
		user := tests.CreateRegularUser(userService)
		perms, err := service.GetUserPermissions(ctx, user.ID)
		assert.Nil(t, err)
		assert.NotNil(t, perms, "GetUserPermissions should never return nil, nil")

		assert.Empty(t, perms, "User without any groups should not have any permissions")
	})

	t.Run("ok: empty permission group", func(t *testing.T) {
		group, err := service.CreatePermissionGroup(ctx, &permissions.PermissionGroup{
			Name:        "test",
			Permissions: nil,
		})
		assert.Nil(t, err)

		group, err = service.GetPermissionGroup(ctx, group.ID)
		assert.Nil(t, err)

		allPermissions, err := service.ListPermissions(ctx)
		assert.Nil(t, err)

		for _, p := range allPermissions {
			enabled, ok := group.Permissions[p]
			assert.True(t, ok, "Could not find permission %q in group permissions", p)
			assert.False(t, enabled, "Permission %q should be disabled in a new, empty group", p)
		}
	})

	t.Run("ok: combined permission groups", func(t *testing.T) {
		// Test data
		user := tests.CreateRegularUser(userService)
		group1, err := service.CreatePermissionGroup(ctx, &permissions.PermissionGroup{
			Name: "group 1",
			Permissions: map[permissions.Permission]bool{
				permissions.PermViewOwnUser: true,
			},
		})
		assert.Nil(t, err)
		group2, err := service.CreatePermissionGroup(ctx, &permissions.PermissionGroup{
			Name: "group 2",
			Permissions: map[permissions.Permission]bool{
				permissions.PermViewOwnOrganisation:  true,
				permissions.PermViewAllOrganisations: true,
			},
		})
		assert.Nil(t, err)

		// Add the test user to both groups
		assert.Nil(t, service.AddUserToPermissionGroup(ctx, user.ID, group1.ID))
		assert.Nil(t, service.AddUserToPermissionGroup(ctx, user.ID, group2.ID))

		// Check that the user has the correct combination of permissions
		perms, err := service.GetUserPermissions(ctx, user.ID)
		assert.Nil(t, err)
		assert.NotNil(t, perms, "GetUserPermissions should never return nil, nil")
		assert.NotEmpty(t, perms, "User with non-empty groups should have permissions")
		for _, p := range []permissions.Permission{
			permissions.PermViewOwnUser,
			permissions.PermViewOwnOrganisation,
			permissions.PermViewAllOrganisations,
		} {
			assert.True(t, perms[p], "Permission %q should be true in combination", p)
		}
	})

	t.Run("ok: existing enabled permission", func(t *testing.T) {
		user := CreateUserWithPermissions(t, db, map[permissions.Permission]bool{
			permissions.PermViewOwnUser:  true,
			permissions.PermEditOwnUser:  true,
			permissions.PermEditAllUsers: false,
		})
		result, err := service.HasAny(ctx, user.ID, permissions.PermViewOwnUser)
		assert.Nil(t, err)
		assert.True(t, result, "Permission that was set to true should return true")
	})

	t.Run("ok: existing disabled permission", func(t *testing.T) {
		user := CreateUserWithPermissions(t, db, map[permissions.Permission]bool{
			permissions.PermViewOwnUser:  true,
			permissions.PermEditOwnUser:  true,
			permissions.PermEditAllUsers: false,
		})
		result, err := service.HasAny(ctx, user.ID, permissions.PermEditAllUsers)
		assert.Nil(t, err)
		assert.False(t, result, "Permission that was set to false should return false")
	})

	t.Run("ok: missing permission should be false", func(t *testing.T) {
		user := CreateUserWithPermissions(t, db, map[permissions.Permission]bool{
			permissions.PermViewOwnUser:  true,
			permissions.PermEditOwnUser:  true,
			permissions.PermEditAllUsers: false,
		})
		result, err := service.HasAny(ctx, user.ID, permissions.PermEditAllOrganisations)
		assert.Nil(t, err)
		assert.False(t, result, "Permission that was not set should return false")
	})

	t.Run("ok: non-existent permission should be false", func(t *testing.T) {
		user := CreateUserWithPermissions(t, db, map[permissions.Permission]bool{
			permissions.PermViewOwnUser:  true,
			permissions.PermEditOwnUser:  true,
			permissions.PermEditAllUsers: false,
		})
		result, err := service.HasAny(ctx, user.ID, permissions.Permission("i do not exist"))
		assert.Nil(t, err)
		assert.False(t, result, "Permission that does not exist should return false")
	})

	t.Run("ok: index is updated after creating a group with a hard-coded id", func(t *testing.T) {
		ctx := context.Background()
		service := postgres.NewPermissionService(db)

		autoGroup1, err := service.CreatePermissionGroup(ctx, &permissions.PermissionGroup{})
		assert.Nil(t, err)

		fixedGroup, err := service.CreatePermissionGroup(ctx, &permissions.PermissionGroup{
			ID: autoGroup1.ID + 12,
		})
		assert.Nil(t, err)

		autoGroup2, err := service.CreatePermissionGroup(ctx, &permissions.PermissionGroup{})
		assert.Nil(t, err)

		assert.Greater(t, fixedGroup.ID, autoGroup1.ID)
		assert.Greater(t, autoGroup2.ID, fixedGroup.ID)
	})
}
