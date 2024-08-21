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
	GetPermissionGroup(ctx context.Context, ID PermissionGroupID) (*PermissionGroup, error)
	// Update a permission group
	UpdatePermissionGroup(ctx context.Context, Group *PermissionGroup) error
	// Delete a permission group
	DeletePermissionGroup(ctx context.Context, ID PermissionGroupID) error
	// Create a new permission group. If no ID was provided, the returned permissiongroup will contain the generated id
	// If an ID was provided as input, the permission group will have that ID. If another group with the same
	// id already exists, this will return core.ErrConflict.
	CreatePermissionGroup(ctx context.Context, Group *PermissionGroup) (*PermissionGroup, error)
	// Rename the specified permission group
	RenamePermissionGroup(ctx context.Context, ID PermissionGroupID, Name string) error
	// Returns whether or not the specified user has the specified permission in any of its permission groups.
	HasAny(ctx context.Context, UserID core.UserID, permission Permission) (bool, error)
	// Returns whether or not the specified user has the specified permission in any of its permission groups for the specified organisation.
	HasAnyForOrg(ctx context.Context, UserID core.UserID, OrgID core.OrganisationID, permission Permission) (bool, error)
	// Lists all permission groups in the system
	ListPermissionGroups(ctx context.Context) ([]PermissionGroup, error)
	// Lists all permission groups for the specified user
	ListPermissionGroupsForUser(ctx context.Context, UserID core.UserID) ([]PermissionGroup, error)
	// Add an existing user to an existing permission group
	AddUserToPermissionGroup(
		ctx context.Context,
		UserID core.UserID,
		GroupID PermissionGroupID,
	) error
	// Return the combined permissions for the specified user.
	// If a user has multiple permission groups, the combined permission group will contain all permissions that are
	// enabled in at least one of their permission groups.
	GetUserPermissions(
		ctx context.Context,
		UserID core.UserID,
	) (map[Permission]bool, error)
	// Lists all permission groups for the specified user in the specified organisation
	ListPermissionGroupsForUserForOrganisation(ctx context.Context, UserID core.UserID, OrgID core.OrganisationID) ([]PermissionGroup, error)
	// Add an existing user to an existing permission group in the specified organisation
	AddUserToPermissionGroupForOrganisation(
		ctx context.Context,
		UserID core.UserID,
		OrgID core.OrganisationID,
		GroupID PermissionGroupID,
	) error
	// Return the combined permissions for the specified user in the specified organisation.
	// If a user has multiple permission groups, the combined permission group will contain all permissions that are
	// enabled in at least one of their permission groups.
	GetUserPermissionsForOrganisation(
		ctx context.Context,
		UserID core.UserID,
		OrgID core.OrganisationID,
	) (map[Permission]bool, error)
}
