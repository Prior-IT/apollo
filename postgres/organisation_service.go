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
) (*core.Organisation, error) {
	return o.AddOrganisation(ctx, o.db, name)
}

// Calls CreateOrganisation query using as a regular query or as a transaction
func (o *OrganisationService) AddOrganisation(
	ctx context.Context,
	dbtx sqlc.DBTX,
	name string,
) (*core.Organisation, error) {
	queries := sqlc.New(dbtx)
	organisation, err := queries.CreateOrganisation(ctx, name)
	if err != nil {
		return nil, convertPgError(err)
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
		return 0, convertPgError(err)
	}
	return uint64(amount), nil
}

// GetOrganisation implements core.OrganisationService.GetOrganisation
func (o *OrganisationService) GetOrganisation(ctx context.Context, id core.OrganisationID) (*core.Organisation, error) {
	organisation, err := o.q.GetOrganisation(ctx, int32(id))
	if err != nil {
		return nil, convertPgError(err)
	}
	return convertOrganisation(organisation)
}

// ListOrganisations implements core.OrganisationService.ListOrganisations
func (o *OrganisationService) ListOrganisations(ctx context.Context) ([]core.Organisation, error) {
	organisations, err := o.q.ListOrganisations(ctx)
	if err != nil {
		return nil, convertPgError(err)
	}
	return convertOrganisationList(organisations)
}

// ListUsersInOrganisation implements core.OrganisationService.ListUsersInOrganisation
func (o *OrganisationService) ListUsersInOrganisation(ctx context.Context, id core.OrganisationID) ([]core.User, error) {
	users, err := o.q.ListUsersInOrganisation(ctx, int32(id))
	if err != nil {
		return nil, convertPgError(err)
	}
	return convertUserList(users)
}

// ListOrganisationsForUser implements core.OrganisationService.ListOrganisationsForUser
func (o *OrganisationService) ListOrganisationsForUser(ctx context.Context, id core.UserID) ([]core.Organisation, error) {
	organisations, err := o.q.ListOrganisationsForUser(ctx, int32(id))
	if err != nil {
		return nil, convertPgError(err)
	}
	return convertOrganisationList(organisations)
}

// AddUserToOrganisation implements core.OrganisationService.AddUserToOrganisation
func (o *OrganisationService) AddUserToOrganisation(ctx context.Context, UserID core.UserID, OrgID core.OrganisationID) error {
	return o.AddUser(ctx, o.db, UserID, OrgID)
}

func (o *OrganisationService) AddUser(ctx context.Context, dbtx sqlc.DBTX, UserID core.UserID, OrgID core.OrganisationID) error {
	queries := sqlc.New(dbtx)
	return queries.AddUserToOrganisation(ctx, int32(UserID), int32(OrgID))
}

// RemoveUserFromOrganisation implements core.OrganisationService.RemoveUserFromOrganisation
func (o *OrganisationService) RemoveUserFromOrganisation(ctx context.Context, UserID core.UserID, OrgID core.OrganisationID) error {
	return o.q.RemoveUserFromOrganisation(ctx, int32(UserID), int32(OrgID))
}

func convertOrganisation(organisation sqlc.ApolloOrganisation) (*core.Organisation, error) {
	id, err := core.NewOrganisationID(uint(organisation.ID))
	if err != nil {
		return nil, err
	}
	return &core.Organisation{
		ID:   id,
		Name: organisation.Name,
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
