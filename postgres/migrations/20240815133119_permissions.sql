-- +goose Up
-- +goose StatementBegin
ALTER TABLE apollo.users
    ADD COLUMN admin boolean NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS apollo.permissions (
    name text NOT NULL,
    PRIMARY KEY (name)
);

CREATE TABLE IF NOT EXISTS apollo.permissiongroups (
    id serial NOT NULL,
    name text,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS apollo.permissiongroup_permissions (
    group_id serial NOT NULL,
    permission text NOT NULL,
    enabled boolean NOT NULL DEFAULT FALSE,
    PRIMARY KEY (group_id, permission),
    FOREIGN KEY (permission) REFERENCES apollo.permissions (name) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES apollo.permissiongroups (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS apollo.user_permissiongroup_membership (
    group_id serial NOT NULL,
    user_id serial NOT NULL,
    PRIMARY KEY (group_id, user_id),
    FOREIGN KEY (user_id) REFERENCES apollo.users (id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES apollo.permissiongroups (id) ON DELETE CASCADE
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS apollo.user_permissiongroup_membership;

DROP TABLE IF EXISTS apollo.permissiongroup_permissions;

DROP TABLE IF EXISTS apollo.permissiongroups;

DROP TABLE IF EXISTS apollo.permissions;

ALTER TABLE apollo.users
    DROP COLUMN admin;

-- +goose StatementEnd
