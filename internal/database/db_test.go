package database

import (
	"testing"

	"github.com/a2sh3r/gophermart/internal/config"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestInitDB_InvalidDSN(t *testing.T) {
	cfg := &config.Config{
		DatabaseURI: "invalid://dsn",
	}

	_, err := InitDB(cfg)
	assert.Error(t, err)
}

func TestInitDB_InvalidMigrationsPath(t *testing.T) {
	cfg := &config.Config{
		DatabaseURI: "postgres://postgres:postgres@localhost:5432/gophermart?sslmode=disable",
	}

	_, err := InitDB(cfg)
	assert.Error(t, err)
}
