-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS apollo.address (
	id serial PRIMARY KEY,
	street text NOT NULL,
	number integer NOT NULL,
	postal_code integer NOT NULL,
	city text NOT NULL,
	country text NOT NULL,
	extra_line text
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS apollo.address;
-- +goose StatementEnd
