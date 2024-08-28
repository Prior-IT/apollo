-- +goose Up
-- +goose StatementBegin
ALTER TABLE apollo.organisations
ADD parent INTEGER NULL,
ADD FOREIGN KEY (parent) REFERENCES apollo.organisations (id) ON DELETE CASCADE;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE apollo.organisations
DROP COLUMN parent;

-- +goose StatementEnd