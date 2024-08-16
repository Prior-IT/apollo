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

-- name: ListPermissionGroupsForUser :many
SELECT
    pg.*
FROM
    apollo.permissiongroups pg
    INNER JOIN apollo.user_permissiongroup_membership usr ON usr.group_id = pg.id
WHERE
    usr.user_id = $1;

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
