package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/prior-it/apollo/core"
	"github.com/prior-it/apollo/permissions"
	"github.com/prior-it/apollo/postgres/internal/sqlc"
)

func NewPermissionService(DB *ApolloDB) *PermissionService {
	sqlc := sqlc.New(DB)
	return &PermissionService{db: DB, q: sqlc}
}

// Postgres implementation of the core UserService interface.
type PermissionService struct {
	db *ApolloDB
	q  *sqlc.Queries
}

// Force struct to implement the interface
var _ permissions.Service = &PermissionService{}

// RegisterPermission implements permissions.Service.
func (p *PermissionService) RegisterPermission(
	ctx context.Context,
	permission permissions.Permission,
) error {
	return p.q.CreatePermission(ctx, permission.String())
}

// ListPermissions implements permissions.Service.
func (p *PermissionService) ListPermissions(ctx context.Context) ([]permissions.Permission, error) {
	dbPermissions, err := p.q.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	perms := make([]permissions.Permission, 0)
	for _, p := range dbPermissions {
		perms = append(perms, permissions.Permission(p))
	}
	return perms, nil
}

// CreatePermissionGroup implements permissions.Service.
func (p *PermissionService) CreatePermissionGroup(
	ctx context.Context,
	Group *permissions.PermissionGroup,
) (*permissions.PermissionGroup, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not create transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // See tx.Rollback() documentation
	q := sqlc.New(tx)

	var NewGroup sqlc.ApolloPermissiongroup
	if Group.ID > 0 {
		NewGroup, err = q.CreatePermissionGroupWithID(ctx, int32(Group.ID), &Group.Name)
		if err != nil {
			return nil, fmt.Errorf(
				"could not create a new permission group with id %v: %w",
				Group.ID,
				convertPgError(err),
			)
		}

	} else {
		NewGroup, err = q.CreatePermissionGroup(ctx, &Group.Name)
		if err != nil {
			return nil, fmt.Errorf("could not create the new permission group: %w", convertPgError(err))
		}
	}

	for permission, enabled := range Group.Permissions {
		err := q.CreatePermissionGroupPermission(ctx, sqlc.CreatePermissionGroupPermissionParams{
			GroupID:    NewGroup.ID,
			Permission: permission.String(),
			Enabled:    enabled,
		})
		if err != nil {
			return nil, fmt.Errorf(
				"could not add permission %q to the new permission group (id %v): %w",
				permission.String(),
				NewGroup.ID,
				err,
			)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("could not commit transaction: %w", err)
	}

	Group.ID = permissions.PermissionGroupID(NewGroup.ID)
	return Group, nil
}

// ListPermissionGroups implements permissions.Service.
func (p *PermissionService) ListPermissionGroups(
	ctx context.Context,
) ([]permissions.PermissionGroup, error) {
	groups, err := p.q.ListPermissionGroups(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]permissions.PermissionGroup, 0)
	for _, g := range groups {
		perms, err := p.q.GetPermissionsForGroup(ctx, g.ID)
		if err != nil {
			return nil, err
		}
		group := combinePermissionGroup(g, perms)
		list = append(list, group)
	}
	return list, nil
}

// ListPermissionGroupsForUser implements permissions.Service.
func (p *PermissionService) ListPermissionGroupsForUser(
	ctx context.Context,
	UserID core.UserID,
) ([]permissions.PermissionGroup, error) {
	groups, err := p.q.ListPermissionGroupsForUser(ctx, int32(UserID))
	if err != nil {
		return nil, err
	}
	list := make([]permissions.PermissionGroup, 0)
	for _, g := range groups {
		perms, err := p.q.GetPermissionsForGroup(ctx, g.ID)
		if err != nil {
			return nil, err
		}
		group := combinePermissionGroup(g, perms)
		list = append(list, group)
	}
	return list, nil
}

// ListPermissionGroupsForUserForOrganisation implements permissions.Service.
func (p *PermissionService) ListPermissionGroupsForUserForOrganisation(
	ctx context.Context,
	UserID core.UserID,
	OrgID core.OrganisationID,
) ([]permissions.PermissionGroup, error) {
	groups, err := p.q.ListPermissionGroupsForUserForOrganisation(ctx, int32(UserID), int32(OrgID))
	if err != nil {
		return nil, err
	}
	list := make([]permissions.PermissionGroup, 0)
	for _, g := range groups {
		perms, err := p.q.GetPermissionsForGroup(ctx, g.ID)
		if err != nil {
			return nil, err
		}
		group := combinePermissionGroup(g, perms)
		list = append(list, group)
	}
	return list, nil
}

// GetPermissionGroup implements permissions.Service.
func (p *PermissionService) GetPermissionGroup(
	ctx context.Context,
	ID permissions.PermissionGroupID,
) (*permissions.PermissionGroup, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not create transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // See tx.Rollback() documentation
	q := sqlc.New(tx)

	group, err := q.GetPermissionGroup(ctx, int32(ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, core.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	permissions, err := q.GetPermissionsForGroup(ctx, group.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("could not commit transaction: %w", err)
	}

	combinedGroup := combinePermissionGroup(group, permissions)
	return &combinedGroup, nil
}

// HasAny implements permissions.Service.
func (p *PermissionService) HasAny(
	ctx context.Context,
	UserID core.UserID,
	permission permissions.Permission,
) (bool, error) {
	groups, err := p.ListPermissionGroupsForUser(ctx, UserID)
	if err != nil {
		return false, err
	}
	for _, g := range groups {
		if g.Get(permission) {
			return true, nil
		}
	}
	return false, nil
}

// HasAnyForOrg implements permissions.Service.
func (p *PermissionService) HasAnyForOrg(
	ctx context.Context,
	UserID core.UserID,
	OrgID core.OrganisationID,
	permission permissions.Permission,
) (bool, error) {
	groups, err := p.ListPermissionGroupsForUserForOrganisation(ctx, UserID, OrgID)
	if err != nil {
		return false, err
	}
	for _, g := range groups {
		if g.Get(permission) {
			return true, nil
		}
	}
	return false, nil
}

// HasAnyForOrgTree implements permissions.Service.
func (p *PermissionService) HasAnyForOrgTree(
	ctx context.Context,
	userID core.UserID,
	orgID core.OrganisationID,
	permission permissions.Permission,
) (bool, error) {
	ok, err := p.HasAnyForOrg(ctx, userID, orgID, permission)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	// Check parent if we didn't find the permission in the current organisation
	parentID, err := p.q.GetParentOrganisation(ctx, int32(orgID))
	if err != nil {
		return false, fmt.Errorf("cannot get parent organisation id: %w", err)
	}
	if parentID != nil {
		pID, err := core.NewOrganisationID(uint(*parentID))
		if err != nil {
			return false, err
		}
		return p.HasAnyForOrgTree(ctx, userID, pID, permission)
	}
	// If there is no more parent organisation and we haven't found it yet, return false
	return false, nil
}

// RenamePermissionGroup implements permissions.Service.
func (p *PermissionService) RenamePermissionGroup(
	ctx context.Context,
	ID permissions.PermissionGroupID,
	Name string,
) error {
	return p.q.RenamePermissionGroup(ctx, int32(ID), &Name)
}

// UpdatePermissionGroup implements permissions.Service.
func (p *PermissionService) UpdatePermissionGroup(
	ctx context.Context,
	Group *permissions.PermissionGroup,
) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("could not create transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // See tx.Rollback() documentation
	q := sqlc.New(tx)
	for permission, enabled := range Group.Permissions {
		err := q.UpdatePermissionGroupPermission(ctx, sqlc.UpdatePermissionGroupPermissionParams{
			GroupID:    int32(Group.ID),
			Permission: permission.String(),
			Enabled:    enabled,
		})
		if err != nil {
			return fmt.Errorf(
				"could not update permission %q in permission group (id %v): %w",
				permission.String(),
				Group.ID,
				err,
			)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}
	return nil
}

// DeletePermissionGroup implements permissions.Service.
func (p *PermissionService) DeletePermissionGroup(
	ctx context.Context,
	GroupID permissions.PermissionGroupID,
) error {
	return p.q.DeletePermissionGroup(ctx, int32(GroupID))
}

// AddUserToPermissionGroup implements permissions.Service.
func (p *PermissionService) AddUserToPermissionGroup(
	ctx context.Context,
	UserID core.UserID,
	GroupID permissions.PermissionGroupID,
) error {
	return p.q.AddUserToPermissionGroup(ctx, int32(GroupID), int32(UserID))
}

// AddUserToPermissionGroup implements permissions.Service.
func (p *PermissionService) AddUserToPermissionGroupForOrganisation(
	ctx context.Context,
	UserID core.UserID,
	OrgID core.OrganisationID,
	GroupID permissions.PermissionGroupID,
) error {
	params := sqlc.AddUserToPermissionGroupForOrganisationParams{
		PermissionGroupID: int32(GroupID),
		UserID:            int32(UserID),
		OrganisationID:    int32(OrgID),
	}
	return p.q.AddUserToPermissionGroupForOrganisation(ctx, params)
}

// GetUserPermissions implements permissions.Service.
func (p *PermissionService) GetUserPermissions(
	ctx context.Context,
	UserID core.UserID,
) (map[permissions.Permission]bool, error) {
	groups, err := p.ListPermissionGroupsForUser(ctx, UserID)
	if err != nil {
		return nil, err
	}

	combined := make(map[permissions.Permission]bool)
	for _, group := range groups {
		for perm, enabled := range group.Permissions {
			combined[perm] = combined[perm] || enabled
		}
	}
	return combined, nil
}

// GetUserPermissionsForOrganisation implements permissions.Service.
func (p *PermissionService) GetUserPermissionsForOrganisation(
	ctx context.Context,
	UserID core.UserID,
	OrgID core.OrganisationID,
) (map[permissions.Permission]bool, error) {
	groups, err := p.ListPermissionGroupsForUserForOrganisation(ctx, UserID, OrgID)
	if err != nil {
		return nil, err
	}

	combined := make(map[permissions.Permission]bool)
	for _, group := range groups {
		for perm, enabled := range group.Permissions {
			combined[perm] = combined[perm] || enabled
		}
	}
	return combined, nil
}

func combinePermissionGroup(
	group sqlc.ApolloPermissiongroup,
	perms []sqlc.GetPermissionsForGroupRow,
) permissions.PermissionGroup {
	Name := ""
	if group.Name != nil {
		Name = *group.Name
	}
	return permissions.PermissionGroup{
		ID:          permissions.PermissionGroupID(group.ID),
		Name:        Name,
		Permissions: cvtPermissions(perms),
	}
}

func cvtPermissions(
	perms []sqlc.GetPermissionsForGroupRow,
) map[permissions.Permission]bool {
	Map := make(map[permissions.Permission]bool)
	for _, perm := range perms {
		Map[permissions.Permission(perm.Permission)] = perm.Enabled
	}
	return Map
}
