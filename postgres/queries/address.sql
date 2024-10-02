-- name: GetAddress :one
SELECT
	*
FROM
	address
WHERE
	id = $1
LIMIT 1;

-- name: CreateAddress :one
INSERT INTO address (street, number, extra_line, postal_code, city, country)
	VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
	*;

-- name: DeleteAddress :exec
DELETE FROM address
WHERE id = $1;

-- name: UpdateAddress :one
UPDATE
	address
SET
	street = COALESCE(sqlc.narg(street), street),
	number = COALESCE(sqlc.narg(number), number),
	extra_line = COALESCE(sqlc.narg(extra_line), extra_line),
	postal_code = COALESCE(sqlc.narg(postal_code), postal_code),
	city = COALESCE(sqlc.narg(city), city),
	country = COALESCE(sqlc.narg(country), country)
WHERE
	id = $1
RETURNING
	*;

-- name: ListAddresses :many
SELECT
	*
FROM
	address;
