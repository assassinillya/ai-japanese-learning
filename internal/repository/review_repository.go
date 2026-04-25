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

func (r *ReviewRepository) ListRecordsByUser(ctx context.Context, userID int64, limit int) ([]model.VocabularyReviewRecordDetail, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT rr.id, rr.user_id, rr.user_vocabulary_id, rr.review_question_id, rr.selected_option, rr.is_correct, rr.reviewed_at,
		       uv.id, uv.user_id, uv.dictionary_entry_id, uv.article_id, uv.source_sentence_id, uv.selected_text, uv.source_sentence_text,
		       uv.status, uv.familiarity, uv.correct_count, uv.wrong_count, uv.consecutive_correct_count, uv.added_at,
		       uv.last_reviewed_at, uv.next_review_at, uv.created_at, uv.updated_at,
		       de.id, de.surface, de.lemma, de.reading, de.romaji, de.part_of_speech, de.meaning_zh, de.meaning_ja, de.meaning_en,
		       de.primary_meaning_zh, de.jlpt_level, de.example_sentence, de.example_translation_zh, de.conjugation_type,
		       de.is_common, de.source, de.verified, de.confidence_score::text, de.ai_model, de.prompt_version, de.created_at, de.updated_at,
		       rq.id, rq.dictionary_entry_id, rq.question_text, rq.correct_answer, rq.option_a, rq.option_b, rq.option_c, rq.option_d,
		       rq.correct_option, rq.explanation_zh, rq.ai_model, rq.prompt_version, rq.created_at,
		       ar.title
		FROM vocabulary_review_records rr
		INNER JOIN user_vocabulary uv ON uv.id = rr.user_vocabulary_id
		INNER JOIN dictionary_entries de ON de.id = uv.dictionary_entry_id
		INNER JOIN vocabulary_review_questions rq ON rq.id = rr.review_question_id
		LEFT JOIN articles ar ON ar.id = uv.article_id
		WHERE rr.user_id = $1
		ORDER BY rr.reviewed_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list review records by user: %w", err)
	}
	defer rows.Close()

	var items []model.VocabularyReviewRecordDetail
	for rows.Next() {
		var item model.VocabularyReviewRecordDetail
		if err := rows.Scan(
			&item.Record.ID,
			&item.Record.UserID,
			&item.Record.UserVocabularyID,
			&item.Record.ReviewQuestionID,
			&item.Record.SelectedOption,
			&item.Record.IsCorrect,
			&item.Record.ReviewedAt,
			&item.UserVocabulary.ID,
			&item.UserVocabulary.UserID,
			&item.UserVocabulary.DictionaryEntryID,
			&item.UserVocabulary.ArticleID,
			&item.UserVocabulary.SourceSentenceID,
			&item.UserVocabulary.SelectedText,
			&item.UserVocabulary.SourceSentenceText,
			&item.UserVocabulary.Status,
			&item.UserVocabulary.Familiarity,
			&item.UserVocabulary.CorrectCount,
			&item.UserVocabulary.WrongCount,
			&item.UserVocabulary.ConsecutiveCorrectCount,
			&item.UserVocabulary.AddedAt,
			&item.UserVocabulary.LastReviewedAt,
			&item.UserVocabulary.NextReviewAt,
			&item.UserVocabulary.CreatedAt,
			&item.UserVocabulary.UpdatedAt,
			&item.Dictionary.ID,
			&item.Dictionary.Surface,
			&item.Dictionary.Lemma,
			&item.Dictionary.Reading,
			&item.Dictionary.Romaji,
			&item.Dictionary.PartOfSpeech,
			&item.Dictionary.MeaningZH,
			&item.Dictionary.MeaningJA,
			&item.Dictionary.MeaningEN,
			&item.Dictionary.PrimaryMeaningZH,
			&item.Dictionary.JLPTLevel,
			&item.Dictionary.ExampleSentence,
			&item.Dictionary.ExampleTranslationZH,
			&item.Dictionary.ConjugationType,
			&item.Dictionary.IsCommon,
			&item.Dictionary.Source,
			&item.Dictionary.Verified,
			&item.Dictionary.ConfidenceScore,
			&item.Dictionary.AIModel,
			&item.Dictionary.PromptVersion,
			&item.Dictionary.CreatedAt,
			&item.Dictionary.UpdatedAt,
			&item.Question.ID,
			&item.Question.DictionaryEntryID,
			&item.Question.QuestionText,
			&item.Question.CorrectAnswer,
			&item.Question.OptionA,
			&item.Question.OptionB,
			&item.Question.OptionC,
			&item.Question.OptionD,
			&item.Question.CorrectOption,
			&item.Question.ExplanationZH,
			&item.Question.AIModel,
			&item.Question.PromptVersion,
			&item.Question.CreatedAt,
			&item.ArticleTitle,
		); err != nil {
			return nil, fmt.Errorf("scan vocabulary review record detail: %w", err)
		}
		item.ContextSentence = item.UserVocabulary.SourceSentenceText
		items = append(items, item)
	}
	return items, rows.Err()
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
