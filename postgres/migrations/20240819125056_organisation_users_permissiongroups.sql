-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS apollo.organisation_users_permissiongroups (
	organisation_users_id serial NOT NULL,
	permission_group_id serial NOT NULL,
	PRIMARY KEY (organisation_users_id, permission_group_id),
	FOREIGN KEY (permission_group_id) REFERENCES apollo.permissiongroups (id) ON DELETE CASCADE,
	FOREIGN KEY (organisation_users_id) REFERENCES apollo.organisation_users (id) ON DELETE CASCADE
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS apollo.organisation_users_permissiongroups;
-- +goose StatementEnd
