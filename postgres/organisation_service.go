package postgres

import (
	"context"
	"errors"

	"github.com/prior-it/apollo/core"
	"github.com/prior-it/apollo/postgres/internal/sqlc"
)

func NewOrganisationService(DB *ApolloDB) *OrganisationService {
	q := sqlc.New(DB)
	return &OrganisationService{DB, q}
}

// Postgres implementation of the core OrganisationService interface.
type OrganisationService struct {
	db *ApolloDB
	q  *sqlc.Queries
}

// Force struct to implement the core interface
var _ core.OrganisationService = &OrganisationService{}

// CreateOrganisation implements core.OrganisationService.CreateOrganisation
func (o *OrganisationService) CreateOrganisation(
	ctx context.Context,
	name string,
	parentID *core.OrganisationID,
) (*core.Organisation, error) {
	return o.AddOrganisation(ctx, o.db, name, parentID)
}

// Calls CreateOrganisation query using as a regular query or as a transaction
func (o *OrganisationService) AddOrganisation(
	ctx context.Context,
	dbtx sqlc.DBTX,
	name string,
	parentID *core.OrganisationID,
) (*core.Organisation, error) {
	queries := sqlc.New(dbtx)
	var castParentID *int32
	if parentID != nil {
		intID := int32(*parentID)
		castParentID = &intID
	}
	organisation, err := queries.CreateOrganisation(ctx, name, castParentID)
	if err != nil {
		return nil, ConvertPgError(err)
	}
	return convertOrganisation(organisation)
}

// DeleteOrganisation implements core.OrganisationService.DeleteOrganisation
func (o *OrganisationService) DeleteOrganisation(ctx context.Context, id core.OrganisationID) error {
	return o.q.DeleteOrganisation(ctx, int32(id))
}

// GetAmountOfOrganisations implements core.OrganisationService.GetAmountOfOrganisations
func (o *OrganisationService) GetAmountOfOrganisations(ctx context.Context) (uint64, error) {
	amount, err := o.q.GetAmountOfOrganisations(ctx)
	if err != nil {
		return 0, ConvertPgError(err)
	}
	return uint64(amount), nil
}

// GetOrganisation implements core.OrganisationService.GetOrganisation
func (o *OrganisationService) GetOrganisation(ctx context.Context, id core.OrganisationID) (*core.Organisation, error) {
	organisation, err := o.q.GetOrganisation(ctx, int32(id))
	if err != nil {
		return nil, ConvertPgError(err)
	}
	return convertOrganisation(organisation)
}

// ListOrganisations implements core.OrganisationService.ListOrganisations
func (o *OrganisationService) ListOrganisations(ctx context.Context) ([]core.Organisation, error) {
	organisations, err := o.q.ListOrganisations(ctx)
	if err != nil {
		return nil, ConvertPgError(err)
	}
	return convertOrganisationList(organisations)
}

// ListOrganisationChildren implements core.OrganisationService.ListOrganisationChildren
func (o *OrganisationService) ListOrganisationChildren(ctx context.Context, parentID core.OrganisationID) ([]core.Organisation, error) {
	i32ParentID := int32(parentID)
	organisations, err := o.q.ListOrganisationChildren(ctx, &i32ParentID)
	if err != nil {
		return nil, ConvertPgError(err)
	}
	return convertOrganisationList(organisations)
}

// ListUsersInOrganisation implements core.OrganisationService.ListUsersInOrganisation
func (o *OrganisationService) ListUsersInOrganisation(ctx context.Context, id core.OrganisationID) ([]core.User, error) {
	users, err := o.q.ListUsersInOrganisation(ctx, int32(id))
	if err != nil {
		return nil, ConvertPgError(err)
	}
	return convertUserList(users)
}

// ListOrganisationsForUser implements core.OrganisationService.ListOrganisationsForUser
func (o *OrganisationService) ListOrganisationsForUser(ctx context.Context, id core.UserID) ([]core.Organisation, error) {
	organisations, err := o.q.ListOrganisationsForUser(ctx, int32(id))
	if err != nil {
		return nil, ConvertPgError(err)
	}
	return convertOrganisationList(organisations)
}

// AddUser implements core.OrganisationService.AddUser
func (o *OrganisationService) AddUser(ctx context.Context, UserID core.UserID, OrgID core.OrganisationID) error {
	return o.AddUserTx(ctx, o.db, UserID, OrgID)
}

func (o *OrganisationService) AddUserTx(ctx context.Context, dbtx sqlc.DBTX, UserID core.UserID, OrgID core.OrganisationID) error {
	queries := sqlc.New(dbtx)
	return queries.AddUserToOrganisation(ctx, int32(UserID), int32(OrgID))
}

// RemoveUser implements core.OrganisationService.RemoveUser
func (o *OrganisationService) RemoveUser(ctx context.Context, UserID core.UserID, OrgID core.OrganisationID) error {
	return o.q.RemoveUserFromOrganisation(ctx, int32(UserID), int32(OrgID))
}

func convertOrganisation(organisation sqlc.ApolloOrganisation) (*core.Organisation, error) {
	id := core.OrganisationID(organisation.ID)
	var parentID *core.OrganisationID
	if organisation.ParentID != nil {
		parentIDVal := core.OrganisationID(*organisation.ParentID)
		parentID = &parentIDVal
	}
	return &core.Organisation{
		ID:       id,
		Name:     organisation.Name,
		ParentID: parentID,
	}, nil
}

func convertOrganisationList(organisations []sqlc.ApolloOrganisation) ([]core.Organisation, error) {
	list := make([]core.Organisation, len(organisations))
	for i, v := range organisations {
		o, err := convertOrganisation(v)
		if err != nil {
			return nil, err
		}
		if o == nil {
			return nil, errors.New("both organisation and error should never be nil")
		}
		list[i] = *o
	}
	return list, nil
}
