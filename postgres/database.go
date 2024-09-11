package postgres

import (
	"context"
	"embed"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

type ApolloDB struct {
	*pgxpool.Pool
}

// Initialise a new database connection. connString should be a valid postgres connection string (such as a postgres-url).
func NewDB(ctx context.Context, connString string) (*ApolloDB, error) {
	slog.Info("Connecting to postgres database", "connString", connString)
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to postgres database: %w", err)
	}
	return &ApolloDB{pool}, nil
}

// Switch the database schema. If the specified schema does not exist already, this will create it.
// Beware that the schema here is not sanitised, as such this could be used to do SQL injection and should never
// pass on unsanitised user input!
func (db *ApolloDB) SwitchSchema(ctx context.Context, schema string) error {
	slog.Info("Switching postgres schema", "schema", schema)
	if _, err := db.Exec(ctx, fmt.Sprintf("BEGIN; SELECT pg_advisory_xact_lock(1); CREATE SCHEMA IF NOT EXISTS %s; COMMIT;", schema)); err != nil {
		return fmt.Errorf("cannot create schema '%v': %w", schema, err)
	}
	if _, err := db.Exec(ctx, fmt.Sprintf("SET search_path TO %s;", schema)); err != nil {
		return fmt.Errorf("cannot set search_path to schema '%v': %w", schema, err)
	}
	return nil
}

// Delete the specified database schema, beware that this will delete all tables and data in the schema.
// The schema string here is not sanitised, as such this could be used to do SQL injection and should never
// pass on unsanitised user input!
func (db *ApolloDB) DeleteSchema(ctx context.Context, schema string) error {
	slog.Info("Deleting postgres schema", "schema", schema)
	if _, err := db.Exec(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE;", schema)); err != nil {
		return fmt.Errorf("cannot delete schema '%v': %w", schema, err)
	}
	return nil
}

// Migrate the database using the specified embedded migration folder.
func (db *ApolloDB) Migrate(folder embed.FS) error {
	goose.SetBaseFS(folder)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("cannot change dialect to postgres: %w", err)
	}

	database := stdlib.OpenDBFromPool(db.Pool)

	if err := goose.Up(database, "migrations"); err != nil {
		return fmt.Errorf("cannot run database migrations: %w", err)
	}

	if err := database.Close(); err != nil {
		return fmt.Errorf("cannot close database connection: %w", err)
	}

	return nil
}

// Migrate the database down a single migration using the specified embedded migration folder.
func (db *ApolloDB) MigrateDown(folder embed.FS) error {
	goose.SetBaseFS(folder)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("cannot change dialect to postgres: %w", err)
	}

	database := stdlib.OpenDBFromPool(db.Pool)

	if err := goose.Down(database, "migrations"); err != nil {
		return fmt.Errorf("cannot run database migrations: %w", err)
	}

	if err := database.Close(); err != nil {
		return fmt.Errorf("cannot close database connection: %w", err)
	}

	return nil
}

// Migrate the Apollo models in the database, if required.
// Note: this will always return the db connection back to the "public" schema,
// use [SwitchSchema] afterwards if you don't want this.
func (db *ApolloDB) MigrateApollo(ctx context.Context) error {
	if err := db.SwitchSchema(ctx, "apollo"); err != nil {
		return fmt.Errorf("cannot switch to the apollo schema: %w", err)
	}

	if err := db.Migrate(migrations); err != nil {
		return fmt.Errorf("cannot run apollo migrations: %w", err)
	}

	if err := db.SwitchSchema(ctx, "public"); err != nil {
		return fmt.Errorf("cannot switch back to the public schema: %w", err)
	}
	return nil
}

// Migrate the Apollo models down a single migration in the database
// Note: this will always return the db connection back to the "public" schema,
// use [SwitchSchema] afterwards if you don't want this.
func (db *ApolloDB) MigrateApolloDown(ctx context.Context) error {
	if err := db.SwitchSchema(ctx, "apollo"); err != nil {
		return fmt.Errorf("cannot switch to the apollo schema: %w", err)
	}

	if err := db.MigrateDown(migrations); err != nil {
		return fmt.Errorf("cannot run apollo migrations: %w", err)
	}

	if err := db.SwitchSchema(ctx, "public"); err != nil {
		return fmt.Errorf("cannot switch back to the public schema: %w", err)
	}
	return nil
}
