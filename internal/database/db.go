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

	logger.Log.Info("Successfully connected to the database", zap.Any("database dsn", cfg.DatabaseURI))
	return db, nil
}

func CloseDB(db *sql.DB) {
	if err := db.Close(); err != nil {
		logger.Log.Error("Database connection closed", zap.Error(err))
	} else {
		logger.Log.Info("Database connection closed")
	}
}
