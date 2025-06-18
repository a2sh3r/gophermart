package config

import (
	"flag"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:"localhost:8080"`
	DatabaseURI          string `env:"DATABASE_URI" envDefault:"postgres://postgres:postgres@localhost:5432/gophermart"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
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
		dbUri          string
		accrualAddress string
		secretKey      string
	)

	flag.StringVar(&runAddress, "a", "", "address host:port")
	flag.StringVar(&dbUri, "d", "", "database host")
	flag.StringVar(&accrualAddress, "r", "", "accrual system host")
	flag.StringVar(&secretKey, "k", "", "secret key to calculate hash")

	flag.Parse()

	if runAddress != "" {
		cfg.RunAddress = runAddress
	}

	if dbUri != "" {
		cfg.DatabaseURI = dbUri
	}

	if accrualAddress != "" {
		cfg.AccrualSystemAddress = accrualAddress
	}

	if secretKey != "" {
		cfg.SecretKey = secretKey
	}

}
