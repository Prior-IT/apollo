-- name: GetUser :one
SELECT
    *
FROM
    apollo.users
WHERE
    id = $1
LIMIT 1;

-- name: ListUsers :many
SELECT
    *
FROM
    apollo.users
ORDER BY
    RANDOM();

-- name: GetAmountOfUsers :one
SELECT
    COUNT(*)
FROM
    apollo.users;

-- name: CreateUser :one
INSERT INTO apollo.users (name, email)
    VALUES ($1, $2)
RETURNING
    *;

-- name: DeleteUser :exec
DELETE FROM apollo.users
WHERE id = $1;

-- name: UpdateUserAdmin :exec
UPDATE
    apollo.users
SET
    admin = $2
WHERE
    id = $1;

-- name: UpdateUser :one
UPDATE
    apollo.users
SET
    name = COALESCE(sqlc.narg(name), name),
    email = COALESCE(sqlc.narg(email), email)
WHERE
    id = $1
RETURNING
    *;
