package postgres

import (
	"context"
	"errors"

	"github.com/prior-it/apollo/core"
	"github.com/prior-it/apollo/postgres/internal/sqlc"
)

func NewUserService(DB *DB) *UserService {
	q := sqlc.New(DB)
	return &UserService{q}
}

// Postgres implementation of the core UserService interface.
type UserService struct {
	q *sqlc.Queries
}

// Force struct to implement the core interface
var _ core.UserService = &UserService{}

// CreateUser implements core.UserService.
func (u *UserService) CreateUser(
	ctx context.Context,
	name string,
	email core.EmailAddress,
) (*core.User, error) {
	if email == nil {
		return nil, errors.New("email cannot be nil")
	}
	user, err := u.q.CreateUser(ctx, name, email.String())
	if err != nil {
		return nil, ConvertPgError(err)
	}
	return convertUser(user)
}

// DeleteUser implements core.UserService.
func (u *UserService) DeleteUser(ctx context.Context, id core.UserID) error {
	return u.q.DeleteUser(ctx, int32(id))
}

// GetAmountOfUsers implements core.UserService.
func (u *UserService) GetAmountOfUsers(ctx context.Context) (uint64, error) {
	amount, err := u.q.GetAmountOfUsers(ctx)
	if err != nil {
		return 0, ConvertPgError(err)
	}
	return uint64(amount), nil
}

// GetUser implements core.UserService.
func (u *UserService) GetUser(ctx context.Context, id core.UserID) (*core.User, error) {
	user, err := u.q.GetUser(ctx, int32(id))
	if err != nil {
		return nil, ConvertPgError(err)
	}
	return convertUser(user)
}

// ListUsers implements core.UserService.
func (u *UserService) ListUsers(ctx context.Context) ([]core.User, error) {
	users, err := u.q.ListUsers(ctx)
	if err != nil {
		return nil, ConvertPgError(err)
	}
	return convertUserList(users)
}

// UpdateUserAdmin implements core.UserService.
func (u *UserService) UpdateUserAdmin(ctx context.Context, id core.UserID, admin bool) error {
	return u.q.UpdateUserAdmin(ctx, int32(id), admin)
}

// UpdateUser implements core.UserService.
func (u *UserService) UpdateUser(
	ctx context.Context,
	id core.UserID,
	data core.UserUpdate,
) (*core.User, error) {
	dbUser, err := u.q.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:    int32(id),
		Name:  data.Name,
		Email: data.Email,
	})
	if err != nil {
		return nil, err
	}
	return convertUser(dbUser)
}

func convertUser(user sqlc.ApolloUser) (*core.User, error) {
	email, err := core.NewEmailAddress(user.Email)
	if err != nil {
		return nil, err
	}
	id := core.UserID(user.ID)
	return &core.User{
		ID:     id,
		Name:   user.Name,
		Email:  email,
		Admin:  user.Admin,
		Joined: user.Joined.Time,
	}, nil
}

func convertUserList(users []sqlc.ApolloUser) ([]core.User, error) {
	list := make([]core.User, len(users))
	for i, v := range users {
		u, err := convertUser(v)
		if err != nil {
			return nil, err
		}
		if u == nil {
			return nil, errors.New("both user and error should never be nil")
		}
		list[i] = *u
	}
	return list, nil
}
