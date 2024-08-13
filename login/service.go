package login

import (
	"context"

	"github.com/prior-it/apollo/core"
)

type UserData struct {
	Name  string `form:"name"        json:"name"`
	Email string `form:"email"       json:"email"`
	// Either an OAuth provider or the login method
	Provider string `form:"provider"    json:"provider"`
	// External user id, currently only used with OAuth
	ProviderID string `form:"provider_id" json:"provider_id"`
}

type Service interface {
	// Return the url that the user should be redirected to to start logging in.
	// Only used in login services that require the user to initiate login (e.g. OAuth)
	GetLoginRedirectURL(provider string, callbackURL string) (string, error)

	// Handle the callback from a login server. The returned UserData might be incomplete if this is a new user and
	// not all data could be retrieved from the chosen provider.
	LoginCallback(
		ctx context.Context,
		provider string,
		code string,
		redirectURL string,
	) (*UserData, error)
}

type AccountService interface {
	// Create a new user and their account.
	CreateUserAccount(
		ctx context.Context,
		data *UserData,
	) (*core.User, error)

	// Find and retrieve a user for the login UserData.
	// If the user does not exist, this will return core.ErrUserDoesNotExist.
	FindUser(
		ctx context.Context,
		data *UserData,
	) (*core.User, error)
}
