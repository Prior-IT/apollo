-- +goose Up
-- +goose StatementBegin
CREATE TYPE user_role AS ENUM (
    'user',
    'admin'
);

CREATE TABLE users (
    id serial NOT NULL,
    name text NOT NULL,
    email text NOT NULL,
    ROLE USER_ROLE NOT NULL DEFAULT 'user',
    joined timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    UNIQUE (email)
);

CREATE TABLE accounts (
    user_id serial NOT NULL,
    provider text NOT NULL,
    provider_id text NOT NULL,
    PRIMARY KEY (user_id, provider),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX unique_provider_provider_id_idx ON accounts (provider, provider_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS unique_provider_provider_id_idx;

DROP TABLE IF EXISTS accounts;

DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS user_role;

-- +goose StatementEnd
