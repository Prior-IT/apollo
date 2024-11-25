-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN lang TEXT NOT NULL DEFAULT 'nl';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
    DROP COLUMN lang;

-- +goose StatementEnd
