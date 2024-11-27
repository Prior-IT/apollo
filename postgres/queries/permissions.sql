-- name: CreatePermission :exec
INSERT INTO permissions (name)
    VALUES ($1)
ON CONFLICT (name)
    DO NOTHING;

-- name: ListPermissions :many
SELECT
    permissions.*
FROM
    permissions;

-- name: ListPermissionGroups :many
SELECT
    permissiongroups.*
FROM
    permissiongroups;

-- name: ListPermissionGroupsForUser :many
SELECT
    pg.*
FROM
    permissiongroups pg
    INNER JOIN user_permissiongroup_membership usr ON usr.group_id = pg.id
WHERE
    usr.user_id = $1;

-- name: ListPermissionGroupsForUserForOrganisation :many
SELECT
    pg.*
FROM
    permissiongroups pg
    INNER JOIN organisation_users_permissiongroups org_usr ON org_usr.permission_group_id = pg.id
WHERE
    org_usr.organisation_users_id = (
        SELECT
            id
        FROM
            organisation_users
        WHERE
            user_id = $1
            AND organisation_id = $2);

-- name: GetPermissionsForGroup :many
SELECT
    p.name AS permission,
    COALESCE(pgp.enabled, FALSE) AS enabled
FROM
    permissions p
    LEFT JOIN permissiongroup_permissions pgp ON p.name = pgp.permission
        AND pgp.group_id = $1;

-- name: GetPermissionGroup :one
SELECT
    pg.*
FROM
    permissiongroups pg
WHERE
    pg.id = $1;

-- name: CreatePermissionGroup :one
INSERT INTO permissiongroups (name)
    VALUES ($1)
RETURNING
    *;

-- name: CreatePermissionGroupWithID :one
INSERT INTO permissiongroups (id, name)
    VALUES ($1, $2)
RETURNING
    *;

-- name: UpdatePermissionGroupIndex :exec
SELECT
    SETVAL('permissiongroups_id_seq', (
            SELECT
                MAX(id)
            FROM permissiongroups));

-- name: CreatePermissionGroupPermission :exec
INSERT INTO permissiongroup_permissions (group_id, permission, enabled)
    VALUES ($1, $2, $3);

-- name: RenamePermissionGroup :exec
UPDATE
    permissiongroups
SET
    name = $2
WHERE
    id = $1;

-- name: UpdatePermissionGroupPermission :exec
UPDATE
    permissiongroup_permissions
SET
    enabled = $3
WHERE
    group_id = $1
    AND permission = $2;

-- name: AddUserToPermissionGroup :exec
INSERT INTO user_permissiongroup_membership (group_id, user_id)
    VALUES ($1, $2);

-- name: AddUserToPermissionGroupForOrganisation :exec
INSERT INTO organisation_users_permissiongroups (permission_group_id, organisation_users_id)
    VALUES ($1, (
            SELECT
                id
            FROM
                organisation_users
            WHERE
                user_id = $2
                AND organisation_id = $3));

-- name: DeletePermissionGroup :exec
DELETE FROM permissiongroups
WHERE permissiongroups.id = $1;
