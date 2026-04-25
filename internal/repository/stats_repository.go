package repository

import (
	"context"
	"database/sql"
	"fmt"

	"ai-japanese-learning/internal/model"
)

type StatsRepository struct {
	db *sql.DB
}

func NewStatsRepository(db *sql.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

func (r *StatsRepository) GetLearningStats(ctx context.Context, userID int64) (*model.LearningStats, error) {
	stats := &model.LearningStats{
		VocabularyStatusCounts: map[string]int64{},
	}

	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM articles
		WHERE user_id = $1
	`, userID).Scan(&stats.ArticleCount); err != nil {
		return nil, fmt.Errorf("count user articles: %w", err)
	}

	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM user_vocabulary
		WHERE user_id = $1
	`, userID).Scan(&stats.VocabularyCount); err != nil {
		return nil, fmt.Errorf("count user vocabulary: %w", err)
	}

	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM user_vocabulary
		WHERE user_id = $1
		  AND status <> 'ignored'
		  AND next_review_at <= NOW()
	`, userID).Scan(&stats.DueVocabularyCount); err != nil {
		return nil, fmt.Errorf("count due vocabulary: %w", err)
	}

	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*),
		       COALESCE(SUM(CASE WHEN is_correct THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN is_correct THEN 0 ELSE 1 END), 0)
		FROM vocabulary_review_records
		WHERE user_id = $1
	`, userID).Scan(&stats.ReviewRecordCount, &stats.ReviewCorrectCount, &stats.ReviewWrongCount); err != nil {
		return nil, fmt.Errorf("count review records: %w", err)
	}

	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*),
		       COALESCE(SUM(CASE WHEN is_correct THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN is_correct THEN 0 ELSE 1 END), 0)
		FROM challenge_question_attempts
		WHERE user_id = $1
	`, userID).Scan(&stats.ReadingAttemptCount, &stats.ReadingCorrectCount, &stats.ReadingWrongCount); err != nil {
		return nil, fmt.Errorf("count reading attempts: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT status, COUNT(*)
		FROM user_vocabulary
		WHERE user_id = $1
		GROUP BY status
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("count vocabulary status: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan vocabulary status count: %w", err)
		}
		stats.VocabularyStatusCounts[status] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}
