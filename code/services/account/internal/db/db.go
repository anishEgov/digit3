package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Init(pool *pgxpool.Pool) {
	Pool = pool
}

func InitDB(pool *pgxpool.Pool) error {
	ctx := context.Background()

	_, err := pool.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS accounts (
		id UUID PRIMARY KEY,
		name TEXT NOT NULL,
		domain TEXT NOT NULL,
		status TEXT NOT NULL CHECK (status IN ('active', 'closed')),
		administrator TEXT NOT NULL,
		oidc_config JSONB NOT NULL,
		created_by TEXT,
		created_on TIMESTAMPTZ,
		modified_by TEXT,
		modified_on TIMESTAMPTZ
	);

	CREATE TABLE IF NOT EXISTS users (
		user_id UUID PRIMARY KEY,
		account_id UUID REFERENCES accounts(id) ON DELETE CASCADE,
		name TEXT,
		email TEXT,
		phone TEXT,
		unique_id TEXT,
		roles JSONB,
		created_by TEXT,
		created_on TIMESTAMPTZ,
		modified_by TEXT,
		modified_on TIMESTAMPTZ
	);

	CREATE TABLE IF NOT EXISTS roles (
		role_id UUID PRIMARY KEY,
		account_id UUID REFERENCES accounts(id) ON DELETE CASCADE,
		name TEXT,
		permissions JSONB,
		created_by TEXT,
		created_on TIMESTAMPTZ,
		modified_by TEXT,
		modified_on TIMESTAMPTZ
	);`)

	return err
}
