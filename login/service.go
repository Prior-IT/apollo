package login

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/prior-it/apollo/core"
)

type UserData struct {
	Name  string `form:"name"        json:"name"`
	Email string `form:"email"       json:"email"`
	Lang  string `form:"lang"        json:"lang"`
	// Either an OAuth provider or the login method
	Provider string `form:"provider"    json:"provider"`
	// External user id, currently only used with OAuth
	ProviderID string `form:"provider_id" json:"provider_id"`
}

type UserDataCacheID struct {
	uuid.UUID
}

func ParseUserDataCacheID(value string) (*UserDataCacheID, error) {
	UUID, err := uuid.Parse(value)
	if err != nil {
		return nil, err
	}
	return &UserDataCacheID{UUID}, nil
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

	// Cache user data so it can be retrieved later and return the cache id.
	CacheUserData(
		ctx context.Context,
		data *UserData,
	) (*UserDataCacheID, error)

	// Return cached user data based on its cache id.
	GetCachedUserData(
		ctx context.Context,
		id *UserDataCacheID,
	) (*UserData, error)

	// Delete all cache entries older than the specified duration.
	DeleteOldCacheEntries(
		ctx context.Context,
		age time.Duration,
	) error

	// Find and retrieve a user for the login UserData.
	// If the user does not exist, this will return core.ErrUserDoesNotExist.
	FindUser(
		ctx context.Context,
		data *UserData,
	) (*core.User, error)
}
