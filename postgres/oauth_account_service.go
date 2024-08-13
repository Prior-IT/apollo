package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
		Email: email.String(),
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

func (s *PgOauthAccountService) CacheUserData(
	ctx context.Context,
	data *login.UserData,
) (*login.UserDataCacheID, error) {
	name := &data.Name
	if len(data.Name) == 0 {
		name = nil
	}
	email := &data.Email
	if len(data.Email) == 0 {
		email = nil
	}
	cache, err := s.q.CreateAccountCache(ctx, sqlc.CreateAccountCacheParams{
		Name:       name,
		Email:      email,
		Provider:   data.Provider,
		ProviderID: data.ProviderID,
	})
	if err != nil {
		return nil, err
	}
	id, err := uuid.FromBytes(cache.ID.Bytes[:])
	if err != nil {
		return nil, err
	}
	cacheID := login.UserDataCacheID{UUID: id}
	return &cacheID, nil
}

func (s *PgOauthAccountService) GetCachedUserData(
	ctx context.Context,
	id *login.UserDataCacheID,
) (*login.UserData, error) {
	var uid pgtype.UUID
	if err := uid.Scan(id.String()); err != nil {
		return nil, err
	}
	cache, err := s.q.GetAccountCacheForID(ctx, uid)
	if err != nil {
		return nil, err
	}

	name := ""
	if cache.Name != nil {
		name = *cache.Name
	}
	email := ""
	if cache.Email != nil {
		email = *cache.Email
	}
	data := login.UserData{
		Name:       name,
		Email:      email,
		Provider:   cache.Provider,
		ProviderID: cache.ProviderID,
	}

	return &data, nil
}

func (s *PgOauthAccountService) DeleteOldCacheEntries(
	ctx context.Context,
	age time.Duration,
) error {
	totalMicroseconds := age.Microseconds()
	days := int32(totalMicroseconds / (24 * time.Hour.Microseconds()))  //nolint:mnd
	microseconds := totalMicroseconds % (24 * time.Hour.Microseconds()) //nolint:mnd
	months := days / 30                                                 //nolint:mnd
	days = days % 30                                                    //nolint:mnd
	interval := pgtype.Interval{
		Microseconds: microseconds,
		Days:         days,
		Months:       months,
		Valid:        true,
	}
	return s.q.DeleteAccountCacheOldEntries(ctx, interval)
}
