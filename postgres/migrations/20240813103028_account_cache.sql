-- +goose Up
-- +goose StatementBegin
CREATE TABLE account_cache (
    id uuid NOT NULL DEFAULT gen_random_uuid (),
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
DROP TABLE IF EXISTS account_cache;

-- +goose StatementEnd
