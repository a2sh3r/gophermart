package config

import (
	"flag"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:"localhost:8084"`
	DatabaseURI          string `env:"DATABASE_URI" envDefault:"postgres://postgres:postgres@localhost:5432/gophermart?sslmode=disable"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://localhost:8080"`
	SecretKey            string `env:"KEY" envDefault:""`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (cfg *Config) ParseFlags() {
	var (
		runAddress     string
		dbURI          string
		accrualAddress string
		secretKey      string
	)

	flag.StringVar(&runAddress, "a", "", "address host:port")
	flag.StringVar(&dbURI, "d", "", "database host")
	flag.StringVar(&accrualAddress, "r", "", "accrual system host")
	flag.StringVar(&secretKey, "k", "", "secret key to calculate hash")

	flag.Parse()

	if runAddress != "" {
		cfg.RunAddress = runAddress
	}

	if dbURI != "" {
		cfg.DatabaseURI = dbURI
	}

	if accrualAddress != "" {
		cfg.AccrualSystemAddress = accrualAddress
	}

	if secretKey != "" {
		cfg.SecretKey = secretKey
	}

}
