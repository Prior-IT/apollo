-- +goose Up
-- +goose StatementBegin
ALTER TABLE apollo.organisations
ADD parent_id INTEGER NULL,
ADD FOREIGN KEY (parent_id) REFERENCES apollo.organisations (id) ON DELETE CASCADE;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE apollo.organisations
DROP COLUMN parent_id;

-- +goose StatementEnd