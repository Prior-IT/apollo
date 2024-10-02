package postgres

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type DB struct {
	*pgxpool.Pool
}

// Initialise a new database connection. connString should be a valid postgres connection string (such as a postgres-url).
func NewDB(ctx context.Context, connString string) (*DB, error) {
	slog.Info("Connecting to postgres database", "connString", connString)
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to postgres database: %w", err)
	}
	return &DB{pool}, nil
}

// Switch the database schema. If the specified schema does not exist already, this will create it.
// Beware that the schema here is not sanitised, as such this could be used to do SQL injection and should never
// pass on unsanitised user input!
func (db *DB) SwitchSchema(ctx context.Context, schema string) error {
	slog.Info("Switching postgres schema", "schema", schema)
	_, err := db.Exec(
		ctx,
		fmt.Sprintf(
			"BEGIN; SELECT pg_advisory_xact_lock(1); CREATE SCHEMA IF NOT EXISTS %s; SET search_path TO %s; COMMIT;",
			schema,
			schema,
		),
	)
	if err != nil {
		return fmt.Errorf("cannot create and switch to schema %q: %w", schema, err)
	}
	return nil
}

// Set the search path to the specified value. If the search path contains non-existent schemas, this will error.
// Beware that the schema here is not sanitised, as such this could be used to do SQL injection and should never
// pass on unsanitised user input!
func (db *DB) SetSearchPath(ctx context.Context, path string) error {
	slog.Info("Changing postgres search path", "search_path", path)
	_, err := db.Exec(
		ctx,
		fmt.Sprintf(
			"BEGIN; SELECT pg_advisory_xact_lock(1); SET search_path TO %s; COMMIT;",
			path,
		),
	)
	if err != nil {
		return fmt.Errorf("cannot set search path to %q: %w", path, err)
	}
	return nil
}

// Delete the specified database schema, beware that this will delete all tables and data in the schema.
// The schema string here is not sanitised, as such this could be used to do SQL injection and should never
// pass on unsanitised user input!
func (db *DB) DeleteSchema(ctx context.Context, schema string) error {
	slog.Info("Deleting postgres schema", "schema", schema)
	if _, err := db.Exec(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE;", schema)); err != nil {
		return fmt.Errorf("cannot delete schema '%v': %w", schema, err)
	}
	return nil
}

// Migrate the database using the specified embedded migration folder.
// "folder" specifies the location of the folder containing sql files within the embed.FS
// To only run the Apollo migrations, set migrations to nil
func (db *DB) Migrate(migrations *embed.FS, folder string) error {
	apollo, err := fs.Sub(embedMigrations, "migrations")
	if err != nil {
		return fmt.Errorf("Cannot get apollo embedFS migrations folder: %w", err)
	}

	migrateFS := apollo

	if migrations != nil {
		app, err := fs.Sub(migrations, folder)
		if err != nil {
			return fmt.Errorf("Cannot get app embedFS migrations folder: %w", err)
		}

		combinedFS := combinedFS{
			fs1: apollo,
			fs2: app,
		}

		migrateFS = combinedFS

	}

	sessionLocker, err := lock.NewPostgresSessionLocker(
		// Timeout after 30min. Try every 15s up to 120 times.
		lock.WithLockTimeout(15, 120), //nolint:mnd
	)
	if err != nil {
		return fmt.Errorf("Cannot use session lock: %w", err)
	}

	database := stdlib.OpenDBFromPool(db.Pool)

	// Create custom goose provider
	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		database,
		migrateFS,
		goose.WithSessionLocker(sessionLocker), // Use session-level advisory lock.
		goose.WithVerbose(true),                // Enable logging (as with goose.Up)
	)
	if err != nil {
		return fmt.Errorf("Cannot create goose provider: %w", err)
	}

	_, err = provider.Up(context.Background())
	if err != nil {
		return fmt.Errorf("cannot run database migrations: %w", err)
	}

	if err := database.Close(); err != nil {
		return fmt.Errorf("cannot close database connection: %w", err)
	}

	return nil
}

// Migrate the database down a single step using the specified embedded migration folder.
// "folder" specifies the location of the folder containing sql files within the embed.FS
// To only run the Apollo migrations, set migrations to nil
func (db *DB) MigrateDown(migrations *embed.FS, folder string) error {
	apollo, err := fs.Sub(embedMigrations, "migrations")
	if err != nil {
		return fmt.Errorf("Cannot get apollo embedFS migrations folder: %w", err)
	}

	migrateFS := apollo

	if migrations != nil {

		app, err := fs.Sub(migrations, folder)
		if err != nil {
			return fmt.Errorf("Cannot get app embedFS migrations folder: %w", err)
		}

		combinedFS := combinedFS{
			fs1: apollo,
			fs2: app,
		}

		migrateFS = combinedFS
	}

	sessionLocker, err := lock.NewPostgresSessionLocker(
		// Timeout after 30min. Try every 15s up to 120 times.
		lock.WithLockTimeout(15, 120), //nolint:mnd
	)
	if err != nil {
		return fmt.Errorf("Cannot use session lock: %w", err)
	}

	database := stdlib.OpenDBFromPool(db.Pool)

	// Create custom goose provider
	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		database,
		migrateFS,
		goose.WithSessionLocker(sessionLocker), // Use session-level advisory lock.
		goose.WithVerbose(true),                // Enable logging (as with goose.Up)
	)
	if err != nil {
		return fmt.Errorf("Cannot create goose provider: %w", err)
	}

	_, err = provider.Down(context.Background())
	if err != nil {
		return fmt.Errorf("cannot run database down migrations: %w", err)
	}

	if err := database.Close(); err != nil {
		return fmt.Errorf("cannot close database connection: %w", err)
	}

	return nil
}

// combinedFS is a custom filesystem that combines two embed.FS instances into a single, larger filesystem
type combinedFS struct {
	fs1, fs2 fs.FS
}

func (c combinedFS) Open(name string) (fs.File, error) {
	f, err := c.fs1.Open(name)
	if err == nil {
		return f, nil
	}
	return c.fs2.Open(name)
}

func (c combinedFS) ReadDir(name string) ([]fs.DirEntry, error) {
	entries1, err1 := fs.ReadDir(c.fs1, name)
	entries2, err2 := fs.ReadDir(c.fs2, name)

	if err1 != nil && err2 != nil {
		return nil, fmt.Errorf("failed to read directory from both filesystems: %v, %v", err1, err2)
	}

	return append(entries1, entries2...), nil
}
