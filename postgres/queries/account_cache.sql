-- name: CreateAccountCache :one
INSERT INTO account_cache (name, email, provider, provider_id)
    VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: DeleteAccountCache :exec
DELETE FROM account_cache
WHERE id = $1;

-- name: DeleteAccountCacheOldEntries :exec
DELETE FROM account_cache
WHERE created < NOW() - $1::interval;

-- name: GetAccountCacheForID :one
SELECT
    *
FROM
    account_cache
WHERE
    id = $1
LIMIT 1;
