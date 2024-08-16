package postgres

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/prior-it/apollo/core"
)

// convertPgError will convert known postgres errors to their core variant.
// Unknown or unhandled errors will be returned as-is.
// Converting nil will simply return nil.
func convertPgError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return errors.Join(core.ErrConflict, err)
		default:
			return err
		}
	} else if errors.Is(err, pgx.ErrNoRows) {
		return errors.Join(core.ErrNotFound, err)
	}
	return err
}
