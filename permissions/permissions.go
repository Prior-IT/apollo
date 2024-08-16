package permissions

import (
	"log/slog"
	"maps"
	"strconv"
)

type (
	Permission string
)

func (p Permission) String() string {
	return string(p)
}

type PermissionGroupID uint

type PermissionGroup struct {
	ID          PermissionGroupID
	Name        string
	Permissions map[Permission]bool
}

func (id PermissionGroupID) String() string {
	return strconv.FormatUint(uint64(id), 10)
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
