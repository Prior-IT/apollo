-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN admin boolean NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS permissions (
    name text NOT NULL,
    PRIMARY KEY (name)
);

CREATE TABLE IF NOT EXISTS permissiongroups (
    id serial NOT NULL,
    name text,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS permissiongroup_permissions (
    group_id serial NOT NULL,
    permission text NOT NULL,
    enabled boolean NOT NULL DEFAULT FALSE,
    PRIMARY KEY (group_id, permission),
    FOREIGN KEY (permission) REFERENCES permissions (name) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES permissiongroups (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_permissiongroup_membership (
    group_id serial NOT NULL,
    user_id serial NOT NULL,
    PRIMARY KEY (group_id, user_id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES permissiongroups (id) ON DELETE CASCADE
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_permissiongroup_membership;

DROP TABLE IF EXISTS permissiongroup_permissions;

DROP TABLE IF EXISTS permissiongroups;

DROP TABLE IF EXISTS permissions;

ALTER TABLE users
    DROP COLUMN admin;

-- +goose StatementEnd
