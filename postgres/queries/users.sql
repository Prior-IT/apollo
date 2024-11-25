-- name: GetUser :one
SELECT
    *
FROM
    users
WHERE
    id = $1
LIMIT 1;

-- name: ListUsers :many
SELECT
    *
FROM
    users
ORDER BY
    RANDOM();

-- name: GetAmountOfUsers :one
SELECT
    COUNT(*)
FROM
    users;

-- name: CreateUser :one
INSERT INTO users (name, email, lang)
    VALUES ($1, $2, $3)
RETURNING
    *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: UpdateUserAdmin :exec
UPDATE
    users
SET
    admin = $2
WHERE
    id = $1;

-- name: UpdateUser :one
UPDATE
    users
SET
    name = COALESCE(sqlc.narg (name), name),
    email = COALESCE(sqlc.narg (email), email),
    lang = COALESCE(sqlc.narg (lang), lang)
WHERE
    id = $1
RETURNING
    *;
