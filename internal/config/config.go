package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL       string
	DeliverooEmail    string
	DeliverooPassword string
}

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:       os.Getenv("GOOSE_DBSTRING"),
		DeliverooEmail:    os.Getenv("DELIVEROO_EMAIL"),
		DeliverooPassword: os.Getenv("DELIVEROO_PASSWORD"),
	}

	var missing []string
	if cfg.DatabaseURL == "" {
		missing = append(missing, "GOOSE_DBSTRING")
	}
	if cfg.DeliverooEmail == "" {
		missing = append(missing, "DELIVEROO_EMAIL")
	}
	if cfg.DeliverooPassword == "" {
		missing = append(missing, "DELIVEROO_PASSWORD")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}
