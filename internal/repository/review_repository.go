package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"ai-japanese-learning/internal/model"
)

var ErrReviewQuestionNotFound = errors.New("review question not found")

type ReviewRepository struct {
	db *sql.DB
}

func NewReviewRepository(db *sql.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) GetQuestionByDictionaryEntry(ctx context.Context, entryID int64) (*model.VocabularyReviewQuestion, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, dictionary_entry_id, question_text, correct_answer, option_a, option_b, option_c, option_d,
		       correct_option, explanation_zh, ai_model, prompt_version, created_at
		FROM vocabulary_review_questions
		WHERE dictionary_entry_id = $1
	`, entryID)

	question, err := scanVocabularyReviewQuestion(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReviewQuestionNotFound
		}
		return nil, fmt.Errorf("get review question by dictionary entry: %w", err)
	}
	return question, nil
}

func (r *ReviewRepository) CreateQuestion(ctx context.Context, question *model.VocabularyReviewQuestion) (*model.VocabularyReviewQuestion, error) {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO vocabulary_review_questions (
			dictionary_entry_id, question_text, correct_answer, option_a, option_b, option_c, option_d,
			correct_option, explanation_zh, ai_model, prompt_version
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (dictionary_entry_id) DO UPDATE
		SET question_text = EXCLUDED.question_text,
		    correct_answer = EXCLUDED.correct_answer,
		    option_a = EXCLUDED.option_a,
		    option_b = EXCLUDED.option_b,
		    option_c = EXCLUDED.option_c,
		    option_d = EXCLUDED.option_d,
		    correct_option = EXCLUDED.correct_option,
		    explanation_zh = EXCLUDED.explanation_zh,
		    ai_model = EXCLUDED.ai_model,
		    prompt_version = EXCLUDED.prompt_version,
		    updated_at = NOW()
		RETURNING id, created_at
	`,
		question.DictionaryEntryID,
		question.QuestionText,
		question.CorrectAnswer,
		question.OptionA,
		question.OptionB,
		question.OptionC,
		question.OptionD,
		question.CorrectOption,
		question.ExplanationZH,
		question.AIModel,
		question.PromptVersion,
	).Scan(&question.ID, &question.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create review question: %w", err)
	}
	return question, nil
}

func (r *ReviewRepository) GetQuestionByID(ctx context.Context, questionID int64) (*model.VocabularyReviewQuestion, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, dictionary_entry_id, question_text, correct_answer, option_a, option_b, option_c, option_d,
		       correct_option, explanation_zh, ai_model, prompt_version, created_at
		FROM vocabulary_review_questions
		WHERE id = $1
	`, questionID)

	question, err := scanVocabularyReviewQuestion(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReviewQuestionNotFound
		}
		return nil, fmt.Errorf("get review question by id: %w", err)
	}
	return question, nil
}

func (r *ReviewRepository) CreateRecord(ctx context.Context, record *model.VocabularyReviewRecord) (*model.VocabularyReviewRecord, error) {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO vocabulary_review_records (
			user_id, user_vocabulary_id, review_question_id, selected_option, is_correct
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, reviewed_at
	`, record.UserID, record.UserVocabularyID, record.ReviewQuestionID, record.SelectedOption, record.IsCorrect).
		Scan(&record.ID, &record.ReviewedAt)
	if err != nil {
		return nil, fmt.Errorf("create review record: %w", err)
	}
	return record, nil
}

type vocabularyReviewQuestionScanner interface {
	Scan(dest ...any) error
}

func scanVocabularyReviewQuestion(scanner vocabularyReviewQuestionScanner) (*model.VocabularyReviewQuestion, error) {
	var question model.VocabularyReviewQuestion
	if err := scanner.Scan(
		&question.ID,
		&question.DictionaryEntryID,
		&question.QuestionText,
		&question.CorrectAnswer,
		&question.OptionA,
		&question.OptionB,
		&question.OptionC,
		&question.OptionD,
		&question.CorrectOption,
		&question.ExplanationZH,
		&question.AIModel,
		&question.PromptVersion,
		&question.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &question, nil
}
