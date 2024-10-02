-- +goose Up
-- +goose StatementBegin
ALTER TABLE organisations
    ADD parent_id INTEGER NULL,
    ADD FOREIGN KEY (parent_id) REFERENCES organisations (id) ON DELETE CASCADE;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE organisations
    DROP COLUMN parent_id;

-- +goose StatementEnd
