-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS apollo.organisations (
	id serial PRIMARY KEY,
	name text NOT NULL
);

CREATE TABLE IF NOT EXISTS apollo.organisation_users (
	id serial PRIMARY KEY,
	user_id serial NOT NULL REFERENCES apollo.users(id) ON DELETE CASCADE,
	organisation_id serial NOT NULL REFERENCES apollo.organisations(id) ON DELETE CASCADE,
	UNIQUE (user_id, organisation_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS apollo.organisation_users;
DROP TABLE IF EXISTS apollo.organisations;
-- +goose StatementEnd
