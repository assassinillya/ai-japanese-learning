package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

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

func (r *VocabularyRepository) ListByUser(ctx context.Context, userID int64, status string) ([]model.VocabularyDetail, error) {
	query := `
		SELECT uv.id, uv.user_id, uv.dictionary_entry_id, uv.article_id, uv.source_sentence_id, uv.selected_text, uv.source_sentence_text,
		       uv.status, uv.familiarity, uv.correct_count, uv.wrong_count, uv.consecutive_correct_count, uv.added_at,
		       uv.last_reviewed_at, uv.next_review_at, uv.created_at, uv.updated_at,
		       de.id, de.surface, de.lemma, de.reading, de.romaji, de.part_of_speech, de.meaning_zh, de.meaning_ja, de.meaning_en,
		       de.primary_meaning_zh, de.jlpt_level, de.example_sentence, de.example_translation_zh, de.conjugation_type,
		       de.is_common, de.source, de.verified, de.confidence_score::text, de.ai_model, de.prompt_version, de.created_at, de.updated_at,
		       a.title
		FROM user_vocabulary uv
		JOIN dictionary_entries de ON de.id = uv.dictionary_entry_id
		LEFT JOIN articles a ON a.id = uv.article_id
		WHERE uv.user_id = $1
	`
	args := []any{userID}
	if strings.TrimSpace(status) != "" {
		query += ` AND uv.status = $2`
		args = append(args, status)
	}
	query += ` ORDER BY uv.added_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list vocabulary by user: %w", err)
	}
	defer rows.Close()

	var items []model.VocabularyDetail
	for rows.Next() {
		item, err := scanVocabularyDetail(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *VocabularyRepository) GetDetail(ctx context.Context, userID, vocabularyID int64) (*model.VocabularyDetail, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT uv.id, uv.user_id, uv.dictionary_entry_id, uv.article_id, uv.source_sentence_id, uv.selected_text, uv.source_sentence_text,
		       uv.status, uv.familiarity, uv.correct_count, uv.wrong_count, uv.consecutive_correct_count, uv.added_at,
		       uv.last_reviewed_at, uv.next_review_at, uv.created_at, uv.updated_at,
		       de.id, de.surface, de.lemma, de.reading, de.romaji, de.part_of_speech, de.meaning_zh, de.meaning_ja, de.meaning_en,
		       de.primary_meaning_zh, de.jlpt_level, de.example_sentence, de.example_translation_zh, de.conjugation_type,
		       de.is_common, de.source, de.verified, de.confidence_score::text, de.ai_model, de.prompt_version, de.created_at, de.updated_at,
		       a.title
		FROM user_vocabulary uv
		JOIN dictionary_entries de ON de.id = uv.dictionary_entry_id
		LEFT JOIN articles a ON a.id = uv.article_id
		WHERE uv.user_id = $1 AND uv.id = $2
	`, userID, vocabularyID)

	item, err := scanVocabularyDetail(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVocabularyNotFound
		}
		return nil, fmt.Errorf("get vocabulary detail: %w", err)
	}
	return item, nil
}

func (r *VocabularyRepository) ListDueForReview(ctx context.Context, userID int64, limit int) ([]model.VocabularyDetail, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT uv.id, uv.user_id, uv.dictionary_entry_id, uv.article_id, uv.source_sentence_id, uv.selected_text, uv.source_sentence_text,
		       uv.status, uv.familiarity, uv.correct_count, uv.wrong_count, uv.consecutive_correct_count, uv.added_at,
		       uv.last_reviewed_at, uv.next_review_at, uv.created_at, uv.updated_at,
		       de.id, de.surface, de.lemma, de.reading, de.romaji, de.part_of_speech, de.meaning_zh, de.meaning_ja, de.meaning_en,
		       de.primary_meaning_zh, de.jlpt_level, de.example_sentence, de.example_translation_zh, de.conjugation_type,
		       de.is_common, de.source, de.verified, de.confidence_score::text, de.ai_model, de.prompt_version, de.created_at, de.updated_at,
		       a.title
		FROM user_vocabulary uv
		JOIN dictionary_entries de ON de.id = uv.dictionary_entry_id
		LEFT JOIN articles a ON a.id = uv.article_id
		WHERE uv.user_id = $1
		  AND uv.status NOT IN ('ignored', 'mastered')
		  AND uv.next_review_at <= NOW()
		ORDER BY uv.next_review_at ASC, uv.added_at ASC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list due vocabulary for review: %w", err)
	}
	defer rows.Close()

	var items []model.VocabularyDetail
	for rows.Next() {
		item, err := scanVocabularyDetail(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *VocabularyRepository) UpdateStatus(ctx context.Context, userID, vocabularyID int64, status model.VocabularyStatus) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE user_vocabulary
		SET status = $3,
		    updated_at = NOW()
		WHERE user_id = $1 AND id = $2
	`, userID, vocabularyID, status)
	if err != nil {
		return fmt.Errorf("update vocabulary status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected for update vocabulary status: %w", err)
	}
	if rows == 0 {
		return ErrVocabularyNotFound
	}
	return nil
}

func (r *VocabularyRepository) UpdateStatusBatch(ctx context.Context, userID int64, vocabularyIDs []int64, status model.VocabularyStatus) (int64, error) {
	if len(vocabularyIDs) == 0 {
		return 0, nil
	}
	placeholders := make([]string, 0, len(vocabularyIDs))
	args := []any{userID, status}
	for index, id := range vocabularyIDs {
		placeholders = append(placeholders, "$"+strconv.Itoa(index+3))
		args = append(args, id)
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE user_vocabulary
		SET status = $2,
		    updated_at = NOW()
		WHERE user_id = $1 AND id IN (`+strings.Join(placeholders, ",")+`)
	`, args...)
	if err != nil {
		return 0, fmt.Errorf("batch update vocabulary status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected for batch update vocabulary status: %w", err)
	}
	return rows, nil
}

func (r *VocabularyRepository) UpdateReviewProgress(
	ctx context.Context,
	userID, vocabularyID int64,
	status model.VocabularyStatus,
	familiarity int,
	correctCount int,
	wrongCount int,
	consecutiveCorrectCount int,
	nextReviewAt time.Time,
) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE user_vocabulary
		SET status = $3,
		    familiarity = $4,
		    correct_count = $5,
		    wrong_count = $6,
		    consecutive_correct_count = $7,
		    last_reviewed_at = NOW(),
		    next_review_at = $8,
		    updated_at = NOW()
		WHERE user_id = $1 AND id = $2
	`, userID, vocabularyID, status, familiarity, correctCount, wrongCount, consecutiveCorrectCount, nextReviewAt)
	if err != nil {
		return fmt.Errorf("update vocabulary review progress: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected for update vocabulary review progress: %w", err)
	}
	if rows == 0 {
		return ErrVocabularyNotFound
	}
	return nil
}

func (r *VocabularyRepository) UpdateContext(
	ctx context.Context,
	userID, dictionaryEntryID int64,
	articleID *int64,
	sourceSentenceID *int64,
	selectedText string,
	sourceSentenceText string,
) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE user_vocabulary
		SET article_id = $3,
		    source_sentence_id = $4,
		    selected_text = $5,
		    source_sentence_text = $6,
		    updated_at = NOW()
		WHERE user_id = $1 AND dictionary_entry_id = $2
	`, userID, dictionaryEntryID, articleID, sourceSentenceID, selectedText, sourceSentenceText)
	if err != nil {
		return fmt.Errorf("update vocabulary context: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected for update vocabulary context: %w", err)
	}
	if rows == 0 {
		return ErrVocabularyNotFound
	}
	return nil
}

func (r *VocabularyRepository) Delete(ctx context.Context, userID, vocabularyID int64) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM user_vocabulary
		WHERE user_id = $1 AND id = $2
	`, userID, vocabularyID)
	if err != nil {
		return fmt.Errorf("delete vocabulary: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected for delete vocabulary: %w", err)
	}
	if rows == 0 {
		return ErrVocabularyNotFound
	}
	return nil
}

func (r *VocabularyRepository) DeleteBatch(ctx context.Context, userID int64, vocabularyIDs []int64) (int64, error) {
	if len(vocabularyIDs) == 0 {
		return 0, nil
	}
	placeholders := make([]string, 0, len(vocabularyIDs))
	args := []any{userID}
	for index, id := range vocabularyIDs {
		placeholders = append(placeholders, "$"+strconv.Itoa(index+2))
		args = append(args, id)
	}
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM user_vocabulary
		WHERE user_id = $1 AND id IN (`+strings.Join(placeholders, ",")+`)
	`, args...)
	if err != nil {
		return 0, fmt.Errorf("batch delete vocabulary: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected for batch delete vocabulary: %w", err)
	}
	return rows, nil
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

func scanVocabularyDetail(scanner userVocabularyScanner) (*model.VocabularyDetail, error) {
	var (
		item         model.UserVocabulary
		entry        model.DictionaryEntry
		articleTitle *string
	)
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
		&entry.ID,
		&entry.Surface,
		&entry.Lemma,
		&entry.Reading,
		&entry.Romaji,
		&entry.PartOfSpeech,
		&entry.MeaningZH,
		&entry.MeaningJA,
		&entry.MeaningEN,
		&entry.PrimaryMeaningZH,
		&entry.JLPTLevel,
		&entry.ExampleSentence,
		&entry.ExampleTranslationZH,
		&entry.ConjugationType,
		&entry.IsCommon,
		&entry.Source,
		&entry.Verified,
		&entry.ConfidenceScore,
		&entry.AIModel,
		&entry.PromptVersion,
		&entry.CreatedAt,
		&entry.UpdatedAt,
		&articleTitle,
	); err != nil {
		return nil, err
	}
	return &model.VocabularyDetail{
		Item:            item,
		DictionaryEntry: entry,
		ArticleTitle:    articleTitle,
		ExampleSentence: item.SourceSentenceText,
	}, nil
}
