-- name: CreateAccount :one
INSERT INTO apollo.accounts (user_id, provider, provider_id)
    VALUES ($1, $2, $3)
RETURNING
    *;

-- name: DeleteAccount :exec
DELETE FROM apollo.accounts
WHERE provider = $1 AND provider_id = $2;

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
