package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/a2sh3r/gophermart/internal/config"
	"github.com/a2sh3r/gophermart/internal/logger"
	"go.uber.org/zap"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

	if err := runMigrations(cfg.DatabaseURI); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	logger.Log.Info("Successfully connected to the database", zap.Any("database dsn", cfg.DatabaseURI))
	return db, nil
}

func runMigrations(dsn string) error {
	m, err := migrate.New(
		"file://internal/migrations",
		dsn,
	)
	if err != nil {
		logger.Log.Error("failed to create migrate instance: %v", zap.Error(err))
		return fmt.Errorf("failed to create migrate instance: %v", err)
	}
	defer func(m *migrate.Migrate) {
		err, _ := m.Close()
		if err != nil {
			logger.Log.Error("failed to close during migration: %v", zap.Error(err))
		}
	}(m)

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	logger.Log.Info("Migrations completed successfully")
	return nil
}
