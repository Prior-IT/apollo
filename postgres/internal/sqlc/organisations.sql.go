// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: organisations.sql

package sqlc

import (
	"context"
)

const addUserToOrganisation = `-- name: AddUserToOrganisation :exec
INSERT INTO organisation_users (user_id, organisation_id)
    VALUES ($1, $2)
`

func (q *Queries) AddUserToOrganisation(ctx context.Context, userID int32, organisationID int32) error {
	_, err := q.db.Exec(ctx, addUserToOrganisation, userID, organisationID)
	return err
}

const createOrganisation = `-- name: CreateOrganisation :one
INSERT INTO organisations (name, parent_id)
    VALUES ($1, $2)
RETURNING
    id, name, parent_id
`

func (q *Queries) CreateOrganisation(ctx context.Context, name string, parentID *int32) (Organisation, error) {
	row := q.db.QueryRow(ctx, createOrganisation, name, parentID)
	var i Organisation
	err := row.Scan(&i.ID, &i.Name, &i.ParentID)
	return i, err
}

const deleteOrganisation = `-- name: DeleteOrganisation :exec
DELETE FROM organisations
WHERE id = $1
`

func (q *Queries) DeleteOrganisation(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, deleteOrganisation, id)
	return err
}

const getAmountOfOrganisations = `-- name: GetAmountOfOrganisations :one
SELECT
    COUNT(id)
FROM
    organisations
`

func (q *Queries) GetAmountOfOrganisations(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, getAmountOfOrganisations)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const getOrganisation = `-- name: GetOrganisation :one
SELECT
    id, name, parent_id
FROM
    organisations
WHERE
    id = $1
LIMIT 1
`

func (q *Queries) GetOrganisation(ctx context.Context, id int32) (Organisation, error) {
	row := q.db.QueryRow(ctx, getOrganisation, id)
	var i Organisation
	err := row.Scan(&i.ID, &i.Name, &i.ParentID)
	return i, err
}

const getParentOrganisation = `-- name: GetParentOrganisation :one
SELECT
    parent_id
FROM
    organisations
WHERE
    id = $1
`

func (q *Queries) GetParentOrganisation(ctx context.Context, id int32) (*int32, error) {
	row := q.db.QueryRow(ctx, getParentOrganisation, id)
	var parent_id *int32
	err := row.Scan(&parent_id)
	return parent_id, err
}

const listOrganisationChildren = `-- name: ListOrganisationChildren :many
SELECT
	id, name, parent_id
FROM
	organisations
WHERE
	parent_id = $1
`

func (q *Queries) ListOrganisationChildren(ctx context.Context, parentID *int32) ([]Organisation, error) {
	rows, err := q.db.Query(ctx, listOrganisationChildren, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Organisation
	for rows.Next() {
		var i Organisation
		if err := rows.Scan(&i.ID, &i.Name, &i.ParentID); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listOrganisations = `-- name: ListOrganisations :many
SELECT
    id, name, parent_id
FROM
    organisations
`

func (q *Queries) ListOrganisations(ctx context.Context) ([]Organisation, error) {
	rows, err := q.db.Query(ctx, listOrganisations)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Organisation
	for rows.Next() {
		var i Organisation
		if err := rows.Scan(&i.ID, &i.Name, &i.ParentID); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listOrganisationsForUser = `-- name: ListOrganisationsForUser :many
SELECT
    o.id, o.name, o.parent_id
FROM
    organisations AS o
    INNER JOIN organisation_users AS ou ON o.id = ou.organisation_id
WHERE
    ou.user_id = $1
`

func (q *Queries) ListOrganisationsForUser(ctx context.Context, userID int32) ([]Organisation, error) {
	rows, err := q.db.Query(ctx, listOrganisationsForUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Organisation
	for rows.Next() {
		var i Organisation
		if err := rows.Scan(&i.ID, &i.Name, &i.ParentID); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listUsersInOrganisation = `-- name: ListUsersInOrganisation :many
SELECT
    u.id, u.name, u.email, u.joined, u.admin
FROM
    users AS u
    INNER JOIN organisation_users AS ou ON u.id = ou.user_id
WHERE
    ou.organisation_id = $1
`

func (q *Queries) ListUsersInOrganisation(ctx context.Context, organisationID int32) ([]User, error) {
	rows, err := q.db.Query(ctx, listUsersInOrganisation, organisationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []User
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Email,
			&i.Joined,
			&i.Admin,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const removeUserFromOrganisation = `-- name: RemoveUserFromOrganisation :exec
DELETE FROM organisation_users
WHERE user_id = $1
    AND organisation_id = $2
`

func (q *Queries) RemoveUserFromOrganisation(ctx context.Context, userID int32, organisationID int32) error {
	_, err := q.db.Exec(ctx, removeUserFromOrganisation, userID, organisationID)
	return err
}

const updateOrganisation = `-- name: UpdateOrganisation :one
UPDATE
    organisations
SET
    name = $2
WHERE
    id = $1
RETURNING
	id, name, parent_id
`

func (q *Queries) UpdateOrganisation(ctx context.Context, iD int32, name string) (Organisation, error) {
	row := q.db.QueryRow(ctx, updateOrganisation, iD, name)
	var i Organisation
	err := row.Scan(&i.ID, &i.Name, &i.ParentID)
	return i, err
}
