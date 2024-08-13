-- +goose Up
-- +goose StatementBegin
CREATE TABLE apollo.account_cache (
    id uuid NOT NULL,
    name text,
    email text,
    provider text NOT NULL,
    provider_id text NOT NULL,
    PRIMARY KEY (id)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS apollo.account_cache;

-- +goose StatementEnd
