package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"ai-japanese-learning/internal/model"
)

var ErrVocabularyNotFound = errors.New("vocabulary entry not found")

type VocabularyRepository struct {
	db *sql.DB
}

func NewVocabularyRepository(db *sql.DB) *VocabularyRepository {
	return &VocabularyRepository{db: db}
}

func (r *VocabularyRepository) GetByUserAndDictionaryEntry(ctx context.Context, userID, entryID int64) (*model.UserVocabulary, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, dictionary_entry_id, article_id, source_sentence_id, selected_text, source_sentence_text,
		       status, familiarity, correct_count, wrong_count, consecutive_correct_count, added_at,
		       last_reviewed_at, next_review_at, created_at, updated_at
		FROM user_vocabulary
		WHERE user_id = $1 AND dictionary_entry_id = $2
	`, userID, entryID)

	item, err := scanUserVocabulary(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVocabularyNotFound
		}
		return nil, fmt.Errorf("get vocabulary by user and dictionary entry: %w", err)
	}
	return item, nil
}

func (r *VocabularyRepository) Create(ctx context.Context, item *model.UserVocabulary) (*model.UserVocabulary, error) {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO user_vocabulary (
			user_id, dictionary_entry_id, article_id, source_sentence_id, selected_text, source_sentence_text,
			status, familiarity, correct_count, wrong_count, consecutive_correct_count, last_reviewed_at, next_review_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, added_at, created_at, updated_at
	`,
		item.UserID,
		item.DictionaryEntryID,
		item.ArticleID,
		item.SourceSentenceID,
		item.SelectedText,
		item.SourceSentenceText,
		item.Status,
		item.Familiarity,
		item.CorrectCount,
		item.WrongCount,
		item.ConsecutiveCorrectCount,
		item.LastReviewedAt,
		item.NextReviewAt,
	).Scan(&item.ID, &item.AddedAt, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create vocabulary item: %w", err)
	}
	return item, nil
}

type userVocabularyScanner interface {
	Scan(dest ...any) error
}

func scanUserVocabulary(scanner userVocabularyScanner) (*model.UserVocabulary, error) {
	var item model.UserVocabulary
	if err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&item.DictionaryEntryID,
		&item.ArticleID,
		&item.SourceSentenceID,
		&item.SelectedText,
		&item.SourceSentenceText,
		&item.Status,
		&item.Familiarity,
		&item.CorrectCount,
		&item.WrongCount,
		&item.ConsecutiveCorrectCount,
		&item.AddedAt,
		&item.LastReviewedAt,
		&item.NextReviewAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &item, nil
}
