package database

import (
	"database/sql"
	"fmt"
	"github.com/a2sh3r/gophermart/internal/config"
	"github.com/a2sh3r/gophermart/internal/logger"
	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func InitDB(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DatabaseURI)
	if err != nil {
		return nil, fmt.Errorf("unable to open database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("unable to ping database: %v", err)
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	logger.Log.Info("Successfully connected to the database", zap.Any("database dsn", cfg.DatabaseURI))
	return db, nil
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			login TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			current_balance NUMERIC(12,2) NOT NULL DEFAULT 0,
			withdrawn_balance NUMERIC(12,2) NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			number TEXT PRIMARY KEY,
			status TEXT NOT NULL DEFAULT 'NEW',
			accrual DOUBLE PRECISION,
			uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			user_id BIGINT NOT NULL REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS withdrawals (
			id SERIAL PRIMARY KEY,
			order_number VARCHAR(255) NOT NULL,
			sum NUMERIC(12,2) NOT NULL CHECK (sum > 0),
			processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
			user_id BIGINT NOT NULL REFERENCES users(id)
		)`,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %v", i+1, err)
		}
		logger.Log.Info("Migration executed successfully", zap.Int("migration", i+1))
	}

	return nil
}
