-- name: CreatePermission :exec
INSERT INTO apollo.permissions (name)
    VALUES ($1)
ON CONFLICT (name)
    DO NOTHING;

-- name: ListPermissions :many
SELECT
    permissions.*
FROM
    apollo.permissions;

-- name: ListPermissionGroups :many
SELECT
    permissiongroups.*
FROM
    apollo.permissiongroups;

-- name: ListPermissionGroupsForUser :many
SELECT
    pg.*
FROM
    apollo.permissiongroups pg
    INNER JOIN apollo.user_permissiongroup_membership usr ON usr.group_id = pg.id
WHERE
    usr.user_id = $1;

-- name: ListPermissionGroupsForUserForOrganisation :many
SELECT
    pg.*
FROM
    apollo.permissiongroups pg
    INNER JOIN apollo.organisation_users_permissiongroups org_usr ON org_usr.permission_group_id = pg.id
WHERE
    org_usr.organisation_users_id = (SELECT id FROM apollo.organisation_users WHERE user_id = $1 AND organisation_id = $2);

-- name: GetPermissionsForGroup :many
SELECT
    p.name AS permission,
    COALESCE(pgp.enabled, FALSE) AS enabled
FROM
    apollo.permissions p
    LEFT JOIN apollo.permissiongroup_permissions pgp ON p.name = pgp.permission
        AND pgp.group_id = $1;

-- name: GetPermissionGroup :one
SELECT
    pg.*
FROM
    apollo.permissiongroups pg
WHERE
    pg.id = $1;

-- name: CreatePermissionGroup :one
INSERT INTO apollo.permissiongroups (name)
    VALUES ($1)
RETURNING
    *;

-- name: CreatePermissionGroupWithID :one
INSERT INTO apollo.permissiongroups (id, name)
    VALUES ($1, $2)
RETURNING
    *;

-- name: CreatePermissionGroupPermission :exec
INSERT INTO apollo.permissiongroup_permissions (group_id, permission, enabled)
    VALUES ($1, $2, $3);

-- name: RenamePermissionGroup :exec
UPDATE
    apollo.permissiongroups
SET
    name = $2
WHERE
    id = $1;

-- name: UpdatePermissionGroupPermission :exec
UPDATE
    apollo.permissiongroup_permissions
SET
    enabled = $3
WHERE
    group_id = $1
    AND permission = $2;

-- name: AddUserToPermissionGroup :exec
INSERT INTO apollo.user_permissiongroup_membership (group_id, user_id)
    VALUES ($1, $2);

-- name: AddUserToPermissionGroupForOrganisation :exec
INSERT INTO apollo.organisation_users_permissiongroups (permission_group_id, organisation_users_id)
    VALUES ($1, (SELECT id FROM apollo.organisation_users WHERE user_id = $2 AND organisation_id = $3));

-- name: DeletePermissionGroup :exec
DELETE FROM apollo.permissiongroups
WHERE permissiongroups.id = $1;
