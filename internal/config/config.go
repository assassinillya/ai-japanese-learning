package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type passwordFile struct {
	PGSQL struct {
		IP       string `json:"ip"`
		Password string `json:"password"`
	} `json:"pgsql"`
}

type Config struct {
	ServerAddress string
	DatabaseURL   string
	TokenSecret   string
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerAddress: ":8080",
		TokenSecret:   os.Getenv("APP_TOKEN_SECRET"),
	}

	if cfg.TokenSecret == "" {
		cfg.TokenSecret = "change-me-in-production"
	}

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		cfg.DatabaseURL = dbURL
		return cfg, nil
	}

	content, err := os.ReadFile("password.json")
	if err != nil {
		return nil, fmt.Errorf("read password.json: %w", err)
	}

	var file passwordFile
	if err := json.Unmarshal(content, &file); err != nil {
		return nil, fmt.Errorf("parse password.json: %w", err)
	}

	host := file.PGSQL.IP
	password := file.PGSQL.Password
	if host == "" || password == "" {
		return nil, fmt.Errorf("password.json missing pgsql.ip or pgsql.password")
	}

	cfg.DatabaseURL = fmt.Sprintf("postgres://postgres:%s@%s/japanese_learning?sslmode=disable", password, host)
	return cfg, nil
}
