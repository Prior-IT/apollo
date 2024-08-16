// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: permissions.sql

package sqlc

import (
	"context"
)

const addUserToPermissionGroup = `-- name: AddUserToPermissionGroup :exec
INSERT INTO apollo.user_permissiongroup_membership (group_id, user_id)
    VALUES ($1, $2)
`

func (q *Queries) AddUserToPermissionGroup(ctx context.Context, groupID int32, userID int32) error {
	_, err := q.db.Exec(ctx, addUserToPermissionGroup, groupID, userID)
	return err
}

const createPermission = `-- name: CreatePermission :exec
INSERT INTO apollo.permissions (name)
    VALUES ($1)
ON CONFLICT (name)
    DO NOTHING
`

func (q *Queries) CreatePermission(ctx context.Context, name string) error {
	_, err := q.db.Exec(ctx, createPermission, name)
	return err
}

const createPermissionGroup = `-- name: CreatePermissionGroup :one
INSERT INTO apollo.permissiongroups (name)
    VALUES ($1)
RETURNING
    id, name
`

func (q *Queries) CreatePermissionGroup(ctx context.Context, name *string) (ApolloPermissiongroup, error) {
	row := q.db.QueryRow(ctx, createPermissionGroup, name)
	var i ApolloPermissiongroup
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const createPermissionGroupPermission = `-- name: CreatePermissionGroupPermission :exec
INSERT INTO apollo.permissiongroup_permissions (group_id, permission, enabled)
    VALUES ($1, $2, $3)
`

type CreatePermissionGroupPermissionParams struct {
	GroupID    int32
	Permission string
	Enabled    bool
}

func (q *Queries) CreatePermissionGroupPermission(ctx context.Context, arg CreatePermissionGroupPermissionParams) error {
	_, err := q.db.Exec(ctx, createPermissionGroupPermission, arg.GroupID, arg.Permission, arg.Enabled)
	return err
}

const createPermissionGroupWithID = `-- name: CreatePermissionGroupWithID :one
INSERT INTO apollo.permissiongroups (id, name)
    VALUES ($1, $2)
RETURNING
    id, name
`

func (q *Queries) CreatePermissionGroupWithID(ctx context.Context, iD int32, name *string) (ApolloPermissiongroup, error) {
	row := q.db.QueryRow(ctx, createPermissionGroupWithID, iD, name)
	var i ApolloPermissiongroup
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const getPermissionGroup = `-- name: GetPermissionGroup :one
SELECT
    pg.id, pg.name
FROM
    apollo.permissiongroups pg
WHERE
    pg.id = $1
`

func (q *Queries) GetPermissionGroup(ctx context.Context, id int32) (ApolloPermissiongroup, error) {
	row := q.db.QueryRow(ctx, getPermissionGroup, id)
	var i ApolloPermissiongroup
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const getPermissionsForGroup = `-- name: GetPermissionsForGroup :many
SELECT
    p.name AS permission,
    COALESCE(pgp.enabled, FALSE) AS enabled
FROM
    apollo.permissions p
    LEFT JOIN apollo.permissiongroup_permissions pgp ON p.name = pgp.permission
        AND pgp.group_id = $1
`

type GetPermissionsForGroupRow struct {
	Permission string
	Enabled    bool
}

func (q *Queries) GetPermissionsForGroup(ctx context.Context, groupID int32) ([]GetPermissionsForGroupRow, error) {
	rows, err := q.db.Query(ctx, getPermissionsForGroup, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPermissionsForGroupRow
	for rows.Next() {
		var i GetPermissionsForGroupRow
		if err := rows.Scan(&i.Permission, &i.Enabled); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listPermissionGroupsForUser = `-- name: ListPermissionGroupsForUser :many
SELECT
    pg.id, pg.name
FROM
    apollo.permissiongroups pg
    INNER JOIN apollo.user_permissiongroup_membership usr ON usr.group_id = pg.id
WHERE
    usr.user_id = $1
`

func (q *Queries) ListPermissionGroupsForUser(ctx context.Context, userID int32) ([]ApolloPermissiongroup, error) {
	rows, err := q.db.Query(ctx, listPermissionGroupsForUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ApolloPermissiongroup
	for rows.Next() {
		var i ApolloPermissiongroup
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listPermissions = `-- name: ListPermissions :many
SELECT
    permissions.name
FROM
    apollo.permissions
`

func (q *Queries) ListPermissions(ctx context.Context) ([]string, error) {
	rows, err := q.db.Query(ctx, listPermissions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		items = append(items, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const renamePermissionGroup = `-- name: RenamePermissionGroup :exec
UPDATE
    apollo.permissiongroups
SET
    name = $2
WHERE
    id = $1
`

func (q *Queries) RenamePermissionGroup(ctx context.Context, iD int32, name *string) error {
	_, err := q.db.Exec(ctx, renamePermissionGroup, iD, name)
	return err
}

const updatePermissionGroupPermission = `-- name: UpdatePermissionGroupPermission :exec
UPDATE
    apollo.permissiongroup_permissions
SET
    enabled = $3
WHERE
    group_id = $1
    AND permission = $2
`

type UpdatePermissionGroupPermissionParams struct {
	GroupID    int32
	Permission string
	Enabled    bool
}

func (q *Queries) UpdatePermissionGroupPermission(ctx context.Context, arg UpdatePermissionGroupPermissionParams) error {
	_, err := q.db.Exec(ctx, updatePermissionGroupPermission, arg.GroupID, arg.Permission, arg.Enabled)
	return err
}