package main

import (
	"context"
	"fmt"

	"github.com/alexedwards/argon2id"
	"github.com/jackc/pgx/v5/pgxpool"
)

func connectDatabase(config *Config) (*pgxpool.Pool, *Queries, error) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, config.Database.URL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("failed to ping database: %w", err)
	}

	queries := New(pool)

	// Check schema version
	version, err := queries.GetSchemaVersion(ctx)
	if err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("failed to get schema version: %w", err)
	}

	if version != config.Schema.Version {
		pool.Close()
		return nil, nil, fmt.Errorf("schema version mismatch: expected %d, got %d", config.Schema.Version, version)
	}

	return pool, queries, nil
}

func ensureDefaultAdmin(ctx context.Context, queries *Queries) error {
	count, err := queries.CountAdmins(ctx)
	if err != nil {
		return fmt.Errorf("failed to count admins: %w", err)
	}

	if count == 0 {
		// Create default admin with username "admin" and password "admin"
		hash, err := argon2id.CreateHash("admin", argon2id.DefaultParams)
		if err != nil {
			return fmt.Errorf("failed to hash default admin password: %w", err)
		}

		_, err = queries.CreateAdmin(ctx, CreateAdminParams{
			Username:     "admin",
			PasswordHash: hash,
		})
		if err != nil {
			return fmt.Errorf("failed to create default admin: %w", err)
		}
	}

	return nil
}
