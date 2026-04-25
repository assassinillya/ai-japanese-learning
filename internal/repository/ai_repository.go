package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"ai-japanese-learning/internal/model"
)

var ErrAICacheNotFound = errors.New("ai cache not found")

type AIRepository struct {
	db *sql.DB
}

func NewAIRepository(db *sql.DB) *AIRepository {
	return &AIRepository{db: db}
}

func (r *AIRepository) GetCache(ctx context.Context, cacheKey string) (*model.AICacheEntry, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, cache_key, task_type, input_hash, request_json::text, response_json::text,
		       model_name, prompt_version, created_at, updated_at
		FROM ai_cache
		WHERE cache_key = $1
	`, cacheKey)

	entry, err := scanAICacheEntry(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAICacheNotFound
		}
		return nil, fmt.Errorf("get ai cache: %w", err)
	}
	return entry, nil
}

func (r *AIRepository) UpsertCache(ctx context.Context, entry *model.AICacheEntry) (*model.AICacheEntry, error) {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO ai_cache (
			cache_key, task_type, input_hash, request_json, response_json, model_name, prompt_version
		)
		VALUES ($1, $2, $3, $4::jsonb, $5::jsonb, $6, $7)
		ON CONFLICT (cache_key) DO UPDATE
		SET request_json = EXCLUDED.request_json,
		    response_json = EXCLUDED.response_json,
		    model_name = EXCLUDED.model_name,
		    prompt_version = EXCLUDED.prompt_version,
		    updated_at = NOW()
		RETURNING id, created_at, updated_at
	`,
		entry.CacheKey,
		entry.TaskType,
		entry.InputHash,
		entry.RequestJSON,
		entry.ResponseJSON,
		entry.ModelName,
		entry.PromptVersion,
	).Scan(&entry.ID, &entry.CreatedAt, &entry.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("upsert ai cache: %w", err)
	}
	return entry, nil
}

func (r *AIRepository) CreateLog(ctx context.Context, log *model.AILog) (*model.AILog, error) {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO ai_logs (
			task_type, request_json, response_json, status, error_message, model_name, prompt_version
		)
		VALUES ($1, $2::jsonb, $3::jsonb, $4, $5, $6, $7)
		RETURNING id, created_at
	`,
		log.TaskType,
		log.RequestJSON,
		log.ResponseJSON,
		log.Status,
		log.ErrorMessage,
		log.ModelName,
		log.PromptVersion,
	).Scan(&log.ID, &log.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create ai log: %w", err)
	}
	return log, nil
}

type aiCacheEntryScanner interface {
	Scan(dest ...any) error
}

func scanAICacheEntry(scanner aiCacheEntryScanner) (*model.AICacheEntry, error) {
	var entry model.AICacheEntry
	if err := scanner.Scan(
		&entry.ID,
		&entry.CacheKey,
		&entry.TaskType,
		&entry.InputHash,
		&entry.RequestJSON,
		&entry.ResponseJSON,
		&entry.ModelName,
		&entry.PromptVersion,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &entry, nil
}
