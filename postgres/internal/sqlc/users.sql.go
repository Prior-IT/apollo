// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: users.sql

package sqlc

import (
	"context"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (name, email, lang)
    VALUES ($1, $2, $3)
RETURNING
    id, name, email, joined, admin, lang
`

type CreateUserParams struct {
	Name  string
	Email string
	Lang  string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser, arg.Name, arg.Email, arg.Lang)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.Joined,
		&i.Admin,
		&i.Lang,
	)
	return i, err
}

const deleteUser = `-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1
`

func (q *Queries) DeleteUser(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, deleteUser, id)
	return err
}

const getAmountOfUsers = `-- name: GetAmountOfUsers :one
SELECT
    COUNT(*)
FROM
    users
`

func (q *Queries) GetAmountOfUsers(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, getAmountOfUsers)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const getUser = `-- name: GetUser :one
SELECT
    id, name, email, joined, admin, lang
FROM
    users
WHERE
    id = $1
LIMIT 1
`

func (q *Queries) GetUser(ctx context.Context, id int32) (User, error) {
	row := q.db.QueryRow(ctx, getUser, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.Joined,
		&i.Admin,
		&i.Lang,
	)
	return i, err
}

const listUsers = `-- name: ListUsers :many
SELECT
    id, name, email, joined, admin, lang
FROM
    users
ORDER BY
    RANDOM()
`

func (q *Queries) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := q.db.Query(ctx, listUsers)
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
			&i.Lang,
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

const updateUser = `-- name: UpdateUser :one
UPDATE
    users
SET
    name = COALESCE($2, name),
    email = COALESCE($3, email),
    lang = COALESCE($4, lang)
WHERE
    id = $1
RETURNING
    id, name, email, joined, admin, lang
`

type UpdateUserParams struct {
	ID    int32
	Name  *string
	Email *string
	Lang  *string
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUser,
		arg.ID,
		arg.Name,
		arg.Email,
		arg.Lang,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.Joined,
		&i.Admin,
		&i.Lang,
	)
	return i, err
}

const updateUserAdmin = `-- name: UpdateUserAdmin :exec
UPDATE
    users
SET
    admin = $2
WHERE
    id = $1
`

func (q *Queries) UpdateUserAdmin(ctx context.Context, iD int32, admin bool) error {
	_, err := q.db.Exec(ctx, updateUserAdmin, iD, admin)
	return err
}
