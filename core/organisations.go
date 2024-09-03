package core

import (
	"context"
)

/**
 * DOMAIN
 */

type Organisation struct {
	ID       OrganisationID
	Name     string
	ParentID *OrganisationID
}

type OrganisationID ID

/**
 * APPLICATION
 */

type OrganisationService interface {
	// Create a new organisation with the specified data.
	CreateOrganisation(ctx context.Context, name string, parentID *OrganisationID) (*Organisation, error)
	// Retrieve the organisation with the specified id or ErrOrganisationDoesNotExist if no such organisation exists.
	GetOrganisation(ctx context.Context, id OrganisationID) (*Organisation, error)
	// Retrieve all existing organisations.
	ListOrganisations(ctx context.Context) ([]Organisation, error)
	// Retrieve the amount of existing organisations.
	GetAmountOfOrganisations(ctx context.Context) (uint64, error)
	// Delete the organisation with the specified id or ErrOrganisationDoesNotExist if no such organisation exists.
	DeleteOrganisation(ctx context.Context, id OrganisationID) error
	// List the organisations a user belongs to or ErrUserDoesNotExist if no such user exists
	ListOrganisationsForUser(ctx context.Context, id UserID) ([]Organisation, error)
	// List the users that belong to an organisation or ErrOrganisationDoesNotExist if no such organisation exists
	ListUsersInOrganisation(ctx context.Context, id OrganisationID) ([]User, error)
	// Add user to an existing organisation
	AddUser(ctx context.Context, UserID UserID, OrgID OrganisationID) error
	// Remove user from an organisation
	RemoveUser(ctx context.Context, UserID UserID, OrgID OrganisationID) error
}
