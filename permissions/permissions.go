package permissions

import (
	"log/slog"
	"maps"
)

type (
	Permission string
)

func (p Permission) String() string {
	return string(p)
}

type PermissionGroup struct {
	ID          int
	Name        string
	Permissions map[Permission]bool
}

func (pg *PermissionGroup) Get(permission Permission) bool {
	value, ok := pg.Permissions[permission]
	if !ok {
		slog.Debug("Unknown permission requested", "permission", permission)
	}
	return value
}

func (pg *PermissionGroup) Clone() *PermissionGroup {
	return &PermissionGroup{
		ID:          0,
		Name:        pg.Name + " (clone)",
		Permissions: maps.Clone(pg.Permissions),
	}
}
