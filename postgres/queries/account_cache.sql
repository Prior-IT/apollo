-- name: CreateAccountCache :one
INSERT INTO apollo.account_cache (name, email, provider, provider_id)
    VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: DeleteAccountCache :exec
DELETE FROM apollo.account_cache
WHERE id = $1;

-- name: DeleteAccountCacheOldEntries :exec
DELETE FROM apollo.account_cache
WHERE created < NOW() - $1::interval;

-- name: GetAccountCacheForID :one
SELECT
    *
FROM
    apollo.account_cache
WHERE
    id = $1
LIMIT 1;
