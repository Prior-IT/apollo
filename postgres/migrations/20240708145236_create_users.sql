-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS apollo;

CREATE TABLE apollo.users (
    id serial NOT NULL,
    name text NOT NULL,
    email text NOT NULL,
    joined timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    UNIQUE (email)
);

CREATE TABLE apollo.accounts (
    user_id serial NOT NULL,
    provider text NOT NULL,
    provider_id text NOT NULL,
    PRIMARY KEY (user_id, provider),
    FOREIGN KEY (user_id) REFERENCES apollo.users (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX unique_provider_provider_id_idx ON apollo.accounts (provider, provider_id);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS unique_provider_provider_id_idx;

DROP TABLE IF EXISTS apollo.accounts;

DROP TABLE IF EXISTS apollo.users;

DROP SCHEMA IF EXISTS apollo;

-- +goose StatementEnd
