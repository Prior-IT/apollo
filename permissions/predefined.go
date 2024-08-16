package permissions

import "context"

const (
	// Users
	PermViewAllUsers Permission = "view_all_users"
	PermEditAllUsers Permission = "edit_all_users"
	PermViewOwnUser  Permission = "view_own_user"
	PermEditOwnUser  Permission = "edit_own_user"

	// Organisations
	PermViewAllOrganisations Permission = "view_all_organisations"
	PermEditAllOrganisations Permission = "edit_all_organisations"
	PermViewOwnOrganisation  Permission = "view_own_organisation"
	PermEditOwnOrganisation  Permission = "edit_own_organisation"

	// Permission groups
	PermViewAllPermissionGroups        Permission = "view_all_permissiongroups"
	PermEditAllPermissionGroups        Permission = "edit_all_permissiongroups"
	PermViewOwnPermissionGroups        Permission = "view_own_permissiongroups"
	PermEditOwnPermissionGroups        Permission = "edit_own_permissiongroups"
	PermEditPermissionGroupPermissions Permission = "edit_permissiongroup_permissions"
)

// Keep this up-to-date based on the permissions above
var allPermissions = [...]Permission{
	PermViewAllUsers,
	PermEditAllUsers,
	PermViewOwnUser,
	PermEditOwnUser,

	PermViewAllOrganisations,
	PermEditAllOrganisations,
	PermViewOwnOrganisation,
	PermEditOwnOrganisation,

	PermViewAllPermissionGroups,
	PermEditAllPermissionGroups,
	PermViewOwnPermissionGroups,
	PermEditOwnPermissionGroups,
	PermEditPermissionGroupPermissions,
}

func RegisterApolloPermissions(service Service) error {
	ctx := context.Background()
	for _, p := range allPermissions {
		err := service.RegisterPermission(ctx, p)
		if err != nil {
			return err
		}
	}
	return nil
}
