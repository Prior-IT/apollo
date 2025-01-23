-- name: GetOrganisation :one
SELECT
    *
FROM
    organisations
WHERE
    id = $1;

-- name: ListOrganisations :many
SELECT
    *
FROM
    organisations;

-- name: GetAmountOfOrganisations :one
SELECT
    COUNT(id)
FROM
    organisations;

-- name: CreateOrganisation :one
INSERT INTO organisations(name, parent_id)
    VALUES ($1, $2)
RETURNING
    *;

-- name: UpdateOrganisation :one
UPDATE
    organisations
SET
    name = $2
WHERE
    id = $1
RETURNING
    *;

-- name: DeleteOrganisation :exec
DELETE FROM organisations
WHERE id = $1;

-- name: ListOrganisationChildren :many
SELECT
    *
FROM
    organisations
WHERE
    parent_id = $1;

-- name: ListOrganisationsForUser :many
SELECT
    o.*
FROM
    organisations AS o
    INNER JOIN organisation_users AS ou ON o.id = ou.organisation_id
WHERE
    ou.user_id = $1;

-- name: ListUsersInOrganisation :many
SELECT
    u.*
FROM
    users AS u
    INNER JOIN organisation_users AS ou ON u.id = ou.user_id
WHERE
    ou.organisation_id = $1;

-- name: AddUserToOrganisation :exec
INSERT INTO organisation_users(user_id, organisation_id)
    VALUES ($1, $2);

-- name: RemoveUserFromOrganisation :exec
DELETE FROM organisation_users
WHERE user_id = $1
    AND organisation_id = $2;

-- name: GetParentOrganisation :one
SELECT
    parent_id
FROM
    organisations
WHERE
    id = $1;

-- name: GetMember :one
SELECT
    users.*
FROM
    users
    INNER JOIN organisation_users ON organisation_users.user_id = users.id
WHERE
    organisation_users.organisation_id = $2
    AND users.id = $1;

-- name: GetMemberByEmail :one
SELECT
    users.*
FROM
    users
    INNER JOIN organisation_users ON organisation_users.user_id = users.id
WHERE
    organisation_users.organisation_id = $1
    AND users.email = $2;
