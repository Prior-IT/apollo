-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS address (
    id serial PRIMARY KEY,
    street text NOT NULL,
    number text NOT NULL,
    postal_code text NOT NULL,
    city text NOT NULL,
    country text NOT NULL,
    extra_line text
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS address;

-- +goose StatementEnd
