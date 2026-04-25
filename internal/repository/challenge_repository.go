package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"ai-japanese-learning/internal/model"
)

var ErrChallengeQuestionNotFound = errors.New("challenge question not found")

type ChallengeRepository struct {
	db *sql.DB
}

func NewChallengeRepository(db *sql.DB) *ChallengeRepository {
	return &ChallengeRepository{db: db}
}

func (r *ChallengeRepository) ListByArticle(ctx context.Context, articleID int64) ([]model.ChallengeQuestion, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, article_id, sentence_id, question_type, question_order, sentence_text, masked_sentence,
		       correct_entry_id, correct_answer_text, option_a, option_b, option_c, option_d,
		       correct_option, explanation, jlpt_level, ai_model, prompt_version, created_at
		FROM challenge_questions
		WHERE article_id = $1
		ORDER BY question_order ASC
	`, articleID)
	if err != nil {
		return nil, fmt.Errorf("list challenge questions by article: %w", err)
	}
	defer rows.Close()

	var questions []model.ChallengeQuestion
	for rows.Next() {
		question, err := scanChallengeQuestion(rows)
		if err != nil {
			return nil, err
		}
		questions = append(questions, *question)
	}
	return questions, rows.Err()
}

func (r *ChallengeRepository) GetByID(ctx context.Context, questionID int64) (*model.ChallengeQuestion, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, article_id, sentence_id, question_type, question_order, sentence_text, masked_sentence,
		       correct_entry_id, correct_answer_text, option_a, option_b, option_c, option_d,
		       correct_option, explanation, jlpt_level, ai_model, prompt_version, created_at
		FROM challenge_questions
		WHERE id = $1
	`, questionID)

	question, err := scanChallengeQuestion(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChallengeQuestionNotFound
		}
		return nil, fmt.Errorf("get challenge question: %w", err)
	}
	return question, nil
}

func (r *ChallengeRepository) GetAccessibleByID(ctx context.Context, userID, questionID int64) (*model.ChallengeQuestion, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT q.id, q.article_id, q.sentence_id, q.question_type, q.question_order, q.sentence_text, q.masked_sentence,
		       q.correct_entry_id, q.correct_answer_text, q.option_a, q.option_b, q.option_c, q.option_d,
		       q.correct_option, q.explanation, q.jlpt_level, q.ai_model, q.prompt_version, q.created_at
		FROM challenge_questions q
		INNER JOIN articles a ON a.id = q.article_id
		WHERE q.id = $1
		  AND (a.user_id = $2 OR a.source_type = 'builtin')
	`, questionID, userID)

	question, err := scanChallengeQuestion(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChallengeQuestionNotFound
		}
		return nil, fmt.Errorf("get accessible challenge question: %w", err)
	}
	return question, nil
}

func (r *ChallengeRepository) ReplaceByArticle(ctx context.Context, articleID int64, questions []model.ChallengeQuestion) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin challenge question tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM challenge_questions WHERE article_id = $1`, articleID); err != nil {
		return fmt.Errorf("delete challenge questions: %w", err)
	}

	for idx := range questions {
		q := &questions[idx]
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO challenge_questions (
				article_id, sentence_id, question_type, question_order, sentence_text, masked_sentence,
				correct_entry_id, correct_answer_text, option_a, option_b, option_c, option_d,
				correct_option, explanation, jlpt_level, ai_model, prompt_version
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
			RETURNING id, created_at
		`,
			q.ArticleID,
			q.SentenceID,
			q.QuestionType,
			q.QuestionOrder,
			q.SentenceText,
			q.MaskedSentence,
			q.CorrectEntryID,
			q.CorrectAnswerText,
			q.OptionA,
			q.OptionB,
			q.OptionC,
			q.OptionD,
			q.CorrectOption,
			q.Explanation,
			q.JLPTLevel,
			q.AIModel,
			q.PromptVersion,
		).Scan(&q.ID, &q.CreatedAt); err != nil {
			return fmt.Errorf("insert challenge question: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit challenge question tx: %w", err)
	}
	return nil
}

func (r *ChallengeRepository) CreateAttempt(ctx context.Context, attempt *model.ChallengeQuestionAttempt) (*model.ChallengeQuestionAttempt, error) {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO challenge_question_attempts (question_id, user_id, selected_option, is_correct)
		VALUES ($1, $2, $3, $4)
		RETURNING id, answered_at
	`, attempt.QuestionID, attempt.UserID, attempt.SelectedOption, attempt.IsCorrect).
		Scan(&attempt.ID, &attempt.AnsweredAt)
	if err != nil {
		return nil, fmt.Errorf("create challenge question attempt: %w", err)
	}
	return attempt, nil
}

type challengeQuestionScanner interface {
	Scan(dest ...any) error
}

func scanChallengeQuestion(scanner challengeQuestionScanner) (*model.ChallengeQuestion, error) {
	var q model.ChallengeQuestion
	if err := scanner.Scan(
		&q.ID,
		&q.ArticleID,
		&q.SentenceID,
		&q.QuestionType,
		&q.QuestionOrder,
		&q.SentenceText,
		&q.MaskedSentence,
		&q.CorrectEntryID,
		&q.CorrectAnswerText,
		&q.OptionA,
		&q.OptionB,
		&q.OptionC,
		&q.OptionD,
		&q.CorrectOption,
		&q.Explanation,
		&q.JLPTLevel,
		&q.AIModel,
		&q.PromptVersion,
		&q.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &q, nil
}
