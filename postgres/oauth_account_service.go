package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/prior-it/apollo/core"
	"github.com/prior-it/apollo/login"
	"github.com/prior-it/apollo/postgres/internal/sqlc"
)

func NewOauthAccountService(db *ApolloDB) *PgOauthAccountService {
	q := sqlc.New(db)
	return &PgOauthAccountService{q, db}
}

// Postgres implementation of the core UserService interface.
type PgOauthAccountService struct {
	q  *sqlc.Queries
	db *ApolloDB
}

// Force struct to implement the core interface
var _ login.AccountService = &PgOauthAccountService{}

func (s *PgOauthAccountService) CreateUserAccount(
	ctx context.Context,
	data *login.UserData,
) (*core.User, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // check rollback documentation

	// Create a Queries object with the new transaction
	qtx := s.q.WithTx(tx)

	email, err := core.NewEmailAddress(data.Email)
	if err != nil {
		return nil, err
	}

	user, err := qtx.CreateUser(ctx, sqlc.CreateUserParams{
		Name:  data.Name,
		Email: email.ToString(),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create user: %w", err)
	}

	_, err = qtx.CreateAccount(ctx, sqlc.CreateAccountParams{
		UserID:     user.ID,
		Provider:   data.Provider,
		ProviderID: data.ProviderID,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create account: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return convertUser(user)
}

func (s *PgOauthAccountService) FindUser(
	ctx context.Context,
	data *login.UserData,
) (*core.User, error) {
	user, err := s.q.GetUserForProvider(ctx, sqlc.GetUserForProviderParams{
		Provider:   data.Provider,
		ProviderID: data.ProviderID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, core.ErrUserDoesNotExist
	} else if err != nil {
		return nil, err
	}
	// core.ErrUserDoesNotExist
	return convertUser(user)
}
