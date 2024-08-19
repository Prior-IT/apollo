package postgres

import (
	"context"
	"errors"

	"github.com/prior-it/apollo/core"
	"github.com/prior-it/apollo/postgres/internal/sqlc"
)

func NewOrganisationService(DB *ApolloDB) *OrganisationService {
	q := sqlc.New(DB)
	return &OrganisationService{q}
}

// Postgres implementation of the core OrganisationService interface.
type OrganisationService struct {
	q *sqlc.Queries
}

// Force struct to implement the core interface
var _ core.OrganisationService = &OrganisationService{}

// CreateOrganisation implements core.OrganisationService.CreateOrganisation
func (u *OrganisationService) CreateOrganisation(
	ctx context.Context,
	name string,
) (*core.Organisation, error) {
	organisation, err := u.q.CreateOrganisation(ctx, name)
	if err != nil {
		return nil, convertPgError(err)
	}
	return convertOrganisation(organisation)
}

// DeleteOrganisation implements core.OrganisationService.DeleteOrganisation
func (u *OrganisationService) DeleteOrganisation(ctx context.Context, id core.OrganisationID) error {
	return u.q.DeleteOrganisation(ctx, int32(id))
}

// GetAmountOfOrganisations implements core.OrganisationService.GetAmountOfOrganisations
func (u *OrganisationService) GetAmountOfOrganisations(ctx context.Context) (uint64, error) {
	amount, err := u.q.GetAmountOfOrganisations(ctx)
	if err != nil {
		return 0, convertPgError(err)
	}
	return uint64(amount), nil
}

// GetOrganisation implements core.OrganisationService.GetOrganisation
func (u *OrganisationService) GetOrganisation(ctx context.Context, id core.OrganisationID) (*core.Organisation, error) {
	organisation, err := u.q.GetOrganisation(ctx, int32(id))
	if err != nil {
		return nil, convertPgError(err)
	}
	return convertOrganisation(organisation)
}

// ListOrganisations implements core.OrganisationService.ListOrganisations
func (u *OrganisationService) ListOrganisations(ctx context.Context) ([]core.Organisation, error) {
	organisations, err := u.q.ListOrganisations(ctx)
	if err != nil {
		return nil, convertPgError(err)
	}
	return convertOrganisationList(organisations)
}

// ListUsersInOrganisation implements core.OrganisationService.ListUsersInOrganisation
func (u *OrganisationService) ListUsersInOrganisation(ctx context.Context, id core.OrganisationID) ([]core.User, error) {
	users, err := u.q.ListUsersInOrganisation(ctx, int32(id))
	if err != nil {
		return nil, convertPgError(err)
	}
	return convertUserList(users)
}

// ListOrganisationsForUser implements core.OrganisationService.ListOrganisationsForUser
func (u *OrganisationService) ListOrganisationsForUser(ctx context.Context, id core.UserID) ([]core.Organisation, error) {
	organisations, err := u.q.ListOrganisationsForUser(ctx, int32(id))
	if err != nil {
		return nil, convertPgError(err)
	}
	return convertOrganisationList(organisations)
}

// AddUserToOrganisation implements core.OrganisationService.AddUserToOrganisation
func (u *OrganisationService) AddUserToOrganisation(ctx context.Context, UserID core.UserID, OrgID core.OrganisationID) error {
	return u.q.AddUserToOrganisation(ctx, int32(UserID), int32(OrgID))
}

// RemoveUserFromOrganisation implements core.OrganisationService.RemoveUserFromOrganisation
func (u *OrganisationService) RemoveUserFromOrganisation(ctx context.Context, UserID core.UserID, OrgID core.OrganisationID) error {
	return u.q.RemoveUserFromOrganisation(ctx, int32(UserID), int32(OrgID))
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
		u, err := convertOrganisation(v)
		if err != nil {
			return nil, err
		}
		if u == nil {
			return nil, errors.New("both organisation and error should never be nil")
		}
		list[i] = *u
	}
	return list, nil
}
