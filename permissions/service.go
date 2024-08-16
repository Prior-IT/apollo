package permissions

import (
	"context"

	"github.com/prior-it/apollo/core"
)

type Service interface {
	// Store a new permission, if it doesn't already exist
	RegisterPermission(ctx context.Context, permission Permission) error
	// Lists all permissions that have been registered before
	ListPermissions(ctx context.Context) ([]Permission, error)
	// Return a permission group by its ID.
	// If the group does not exist, this returns core.ErrNotFound
	GetPermissionGroup(ctx context.Context, ID int) (*PermissionGroup, error)
	// Update a permission group
	UpdatePermissionGroup(ctx context.Context, Group *PermissionGroup) error
	// Create a new permission group. If no ID was provided, the returned permissiongroup will contain the generated id
	// If an ID was provided as input, the permission group will have that ID. If another group with the same
	// id already exists, this will return core.ErrConflict.
	CreatePermissionGroup(ctx context.Context, Group *PermissionGroup) (*PermissionGroup, error)
	// Rename the specified permission group
	RenamePermissionGroup(ctx context.Context, ID int, Name string) error
	// Returns whether or not the specified user has the specified permission in any of its permission groups.
	HasAny(ctx context.Context, UserID core.UserID, permission Permission) (bool, error)
	// Lists all permission groups for the specified user
	ListPermissionGroups(ctx context.Context, UserID core.UserID) ([]PermissionGroup, error)
	// Add an existing user to an existing permission group
	AddUserToPermissionGroup(ctx context.Context, UserID core.UserID, GroupID int) error
	// Return the combined permissions for the specified user.
	// If a user has multiple permission groups, the combined permission group will contain all permissions that are
	// enabled in at least one of their permission groups.
	GetUserPermissions(
		ctx context.Context,
		UserID core.UserID,
	) (map[Permission]bool, error)
}
