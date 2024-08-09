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

-- name: GetUserForProvider :one
SELECT
    users.*
FROM
    apollo.users
    INNER JOIN apollo.accounts ON users.id = accounts.user_id
WHERE
    accounts.provider = $1
    AND accounts.provider_id = $2
LIMIT 1;

-- name: DeleteUser :exec
DELETE FROM apollo.users
WHERE id = $1;
