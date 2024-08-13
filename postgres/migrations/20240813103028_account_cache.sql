-- +goose Up
-- +goose StatementBegin
CREATE TABLE apollo.account_cache (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    name text,
    email text,
    provider text NOT NULL,
    provider_id text NOT NULL,
    created timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS apollo.account_cache;

-- +goose StatementEnd
