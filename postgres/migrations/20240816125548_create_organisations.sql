-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS organisations (
    id serial NOT NULL,
    name text NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS organisation_users (
    id serial NOT NULL,
    user_id serial NOT NULL,
    organisation_id serial NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (organisation_id) REFERENCES organisations (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX unique_organisation_id_users_id_idx ON organisation_users (user_id, organisation_id);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS unique_organisation_id_users_id_idx;

DROP TABLE IF EXISTS organisation_users;

DROP TABLE IF EXISTS organisations;

-- +goose StatementEnd
