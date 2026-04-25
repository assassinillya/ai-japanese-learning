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
	return r.ListByArticleAndType(ctx, articleID, "")
}

func (r *ChallengeRepository) ListByArticleAndType(ctx context.Context, articleID int64, questionType string) ([]model.ChallengeQuestion, error) {
	query := `
		SELECT id, article_id, sentence_id, question_type, question_order, sentence_text, masked_sentence,
		       correct_entry_id, correct_answer_text, option_a, option_b, option_c, option_d,
		       correct_option, explanation, jlpt_level, ai_model, prompt_version, created_at
		FROM challenge_questions
		WHERE article_id = $1
	`
	args := []any{articleID}
	if questionType != "" {
		query += ` AND question_type = $2`
		args = append(args, questionType)
	}
	query += ` ORDER BY question_order ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
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
	return r.ReplaceByArticleAndType(ctx, articleID, "", questions)
}

func (r *ChallengeRepository) ReplaceByArticleAndType(ctx context.Context, articleID int64, questionType string, questions []model.ChallengeQuestion) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin challenge question tx: %w", err)
	}
	defer tx.Rollback()

	deleteQuery := `DELETE FROM challenge_questions WHERE article_id = $1`
	deleteArgs := []any{articleID}
	if questionType != "" {
		deleteQuery += ` AND question_type = $2`
		deleteArgs = append(deleteArgs, questionType)
	}
	if _, err := tx.ExecContext(ctx, deleteQuery, deleteArgs...); err != nil {
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
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
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

func (r *ChallengeRepository) ListAttemptsByArticleAndType(ctx context.Context, userID, articleID int64, questionType string) ([]model.ReadingAnswerDetail, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.id, a.question_id, a.user_id, a.selected_option, a.is_correct, a.answered_at,
		       q.id, q.article_id, q.sentence_id, q.question_type, q.question_order, q.sentence_text, q.masked_sentence,
		       q.correct_entry_id, q.correct_answer_text, q.option_a, q.option_b, q.option_c, q.option_d,
		       q.correct_option, q.explanation, q.jlpt_level, q.ai_model, q.prompt_version, q.created_at
		FROM challenge_question_attempts a
		INNER JOIN challenge_questions q ON q.id = a.question_id
		INNER JOIN articles ar ON ar.id = q.article_id
		WHERE a.user_id = $1
		  AND q.article_id = $2
		  AND q.question_type = $3
		  AND (ar.user_id = $1 OR ar.source_type = 'builtin')
		ORDER BY a.answered_at DESC, q.question_order ASC
	`, userID, articleID, questionType)
	if err != nil {
		return nil, fmt.Errorf("list reading attempts by article and type: %w", err)
	}
	defer rows.Close()

	var items []model.ReadingAnswerDetail
	for rows.Next() {
		var item model.ReadingAnswerDetail
		if err := rows.Scan(
			&item.Attempt.ID,
			&item.Attempt.QuestionID,
			&item.Attempt.UserID,
			&item.Attempt.SelectedOption,
			&item.Attempt.IsCorrect,
			&item.Attempt.AnsweredAt,
			&item.Question.ID,
			&item.Question.ArticleID,
			&item.Question.SentenceID,
			&item.Question.QuestionType,
			&item.Question.QuestionOrder,
			&item.Question.SentenceText,
			&item.Question.MaskedSentence,
			&item.Question.CorrectEntryID,
			&item.Question.CorrectAnswerText,
			&item.Question.OptionA,
			&item.Question.OptionB,
			&item.Question.OptionC,
			&item.Question.OptionD,
			&item.Question.CorrectOption,
			&item.Question.Explanation,
			&item.Question.JLPTLevel,
			&item.Question.AIModel,
			&item.Question.PromptVersion,
			&item.Question.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan reading answer detail: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
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
