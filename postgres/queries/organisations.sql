-- name: GetOrganisation :one
SELECT
	*
FROM
	apollo.organisations
WHERE
	id = $1
LIMIT 1;

-- name: ListOrganisations :many
SELECT
	*
FROM
	apollo.organisations;

-- name: GetAmountOfOrganisations :one
SELECT
	COUNT(id)
FROM
	apollo.organisations;

-- name: CreateOrganisation :one
INSERT INTO apollo.organisations (name)
	VALUES ($1)
RETURNING
	*;

-- name: DeleteOrganisation :exec
DELETE FROM apollo.organisations
WHERE id = $1;

-- name: ListOrganisationsForUser :many
SELECT
	o.*
FROM
	apollo.organisations AS o
	INNER JOIN apollo.organisation_users AS ou ON o.id = ou.organisation_id
WHERE
	ou.user_id = $1;

-- name: ListUsersInOrganisation :many
SELECT
	u.*
FROM
	apollo.users AS u
	INNER JOIN apollo.organisation_users AS ou ON u.id = ou.user_id
WHERE
	ou.organisation_id = $1;

-- name: AddUserToOrganisation :exec
INSERT INTO apollo.organisation_users (user_id, organisation_id)
	VALUES ($1, $2);

-- name: RemoveUserFromOrganisation :exec
DELETE FROM
	apollo.organisation_users
WHERE
	user_id = $1 AND organisation_id = $2;
