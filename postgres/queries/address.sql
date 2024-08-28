-- name: GetAddress :one
SELECT
	*
FROM
	apollo.address
WHERE
	id = $1
LIMIT 1;

-- name: CreateAddress :one
INSERT INTO apollo.address (street, number, extra_line, postal_code, city, country)
	VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
	*;

-- name: DeleteAddress :exec
DELETE FROM apollo.address
WHERE id = $1;
