package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrUserAIConfigNotFound = errors.New("user ai config not found")

type UserAIConfigRepository struct {
	db *sql.DB
}

type UserAIConfig struct {
	Provider     string
	ProviderName string
	BaseURL      string
	APIKey       string
	Model        string
	APIVersion   string
}

func NewUserAIConfigRepository(db *sql.DB) *UserAIConfigRepository {
	return &UserAIConfigRepository{db: db}
}

func (r *UserAIConfigRepository) EnsureTable(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS user_ai_configs (
			user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			provider TEXT NOT NULL,
			provider_name TEXT NOT NULL DEFAULT '',
			base_url TEXT NOT NULL DEFAULT '',
			api_key TEXT NOT NULL DEFAULT '',
			model TEXT NOT NULL DEFAULT '',
			api_version TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("ensure user ai config table: %w", err)
	}
	return nil
}

func (r *UserAIConfigRepository) Get(ctx context.Context, userID int64) (UserAIConfig, error) {
	var cfg UserAIConfig
	err := r.db.QueryRowContext(ctx, `
		SELECT provider, provider_name, base_url, api_key, model, api_version
		FROM user_ai_configs
		WHERE user_id = $1
	`, userID).Scan(&cfg.Provider, &cfg.ProviderName, &cfg.BaseURL, &cfg.APIKey, &cfg.Model, &cfg.APIVersion)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserAIConfig{}, ErrUserAIConfigNotFound
		}
		return UserAIConfig{}, fmt.Errorf("get user ai config: %w", err)
	}
	return cfg, nil
}

func (r *UserAIConfigRepository) Upsert(ctx context.Context, userID int64, cfg UserAIConfig) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_ai_configs (user_id, provider, provider_name, base_url, api_key, model, api_version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE SET
			provider = EXCLUDED.provider,
			provider_name = EXCLUDED.provider_name,
			base_url = EXCLUDED.base_url,
			api_key = EXCLUDED.api_key,
			model = EXCLUDED.model,
			api_version = EXCLUDED.api_version,
			updated_at = NOW()
	`, userID, cfg.Provider, cfg.ProviderName, cfg.BaseURL, cfg.APIKey, cfg.Model, cfg.APIVersion)
	if err != nil {
		return fmt.Errorf("upsert user ai config: %w", err)
	}
	return nil
}
