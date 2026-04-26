package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"ai-japanese-learning/internal/model"
)

type ArticleRepository struct {
	db *sql.DB
}

var ErrArticleNotFound = errors.New("article not found")

func NewArticleRepository(db *sql.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) ListLibrary(ctx context.Context) ([]model.Article, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, title, original_language, original_content, japanese_content, chinese_translation,
		       jlpt_level, source_type, is_ai_generated, is_verified, translation_status,
		       processing_notes, sentence_count, created_at, updated_at
		FROM articles
		WHERE source_type = 'builtin'
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list library articles: %w", err)
	}
	defer rows.Close()

	var articles []model.Article
	for rows.Next() {
		article, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		articles = append(articles, *article)
	}
	return articles, rows.Err()
}

func (r *ArticleRepository) ListByUser(ctx context.Context, userID int64) ([]model.Article, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, title, original_language, original_content, japanese_content, chinese_translation,
		       jlpt_level, source_type, is_ai_generated, is_verified, translation_status,
		       processing_notes, sentence_count, created_at, updated_at
		FROM articles
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list user articles: %w", err)
	}
	defer rows.Close()

	var articles []model.Article
	for rows.Next() {
		article, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		articles = append(articles, *article)
	}
	return articles, rows.Err()
}

func (r *ArticleRepository) ListPublic(ctx context.Context) ([]model.Article, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, title, original_language, original_content, japanese_content, chinese_translation,
		       jlpt_level, source_type, is_ai_generated, is_verified, translation_status,
		       processing_notes, sentence_count, created_at, updated_at
		FROM articles
		WHERE translation_status = 'done'
		ORDER BY updated_at DESC, created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list public articles: %w", err)
	}
	defer rows.Close()

	var articles []model.Article
	for rows.Next() {
		article, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		articles = append(articles, *article)
	}
	return articles, rows.Err()
}

func (r *ArticleRepository) GetAccessible(ctx context.Context, userID, articleID int64) (*model.Article, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, title, original_language, original_content, japanese_content, chinese_translation,
		       jlpt_level, source_type, is_ai_generated, is_verified, translation_status,
		       processing_notes, sentence_count, created_at, updated_at
		FROM articles
		WHERE id = $1
	`, articleID)

	article, err := scanArticle(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrArticleNotFound
		}
		return nil, fmt.Errorf("get article: %w", err)
	}
	return article, nil
}

func (r *ArticleRepository) Create(ctx context.Context, article *model.Article) (*model.Article, error) {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO articles (
			user_id, title, original_language, original_content, japanese_content,
			chinese_translation, jlpt_level, source_type, is_ai_generated, is_verified,
			translation_status, processing_notes, sentence_count
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`,
		article.UserID,
		article.Title,
		article.OriginalLanguage,
		article.OriginalContent,
		article.JapaneseContent,
		article.ChineseTranslation,
		article.JLPTLevel,
		article.SourceType,
		article.IsAIGenerated,
		article.IsVerified,
		article.TranslationStatus,
		article.ProcessingNotes,
		article.SentenceCount,
	).Scan(&article.ID, &article.CreatedAt, &article.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create article: %w", err)
	}
	return article, nil
}

func (r *ArticleRepository) UpdateProcessedContent(
	ctx context.Context,
	articleID int64,
	japaneseContent string,
	chineseTranslation *string,
	translationStatus model.TranslationStatus,
	sourceType string,
	isAIGenerated bool,
	processingNotes *string,
	sentenceCount int,
) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE articles
		SET japanese_content = $2,
		    chinese_translation = $3,
		    translation_status = $4,
		    source_type = $5,
		    is_ai_generated = $6,
		    processing_notes = $7,
		    sentence_count = $8,
		    updated_at = NOW()
		WHERE id = $1
	`, articleID, japaneseContent, chineseTranslation, translationStatus, sourceType, isAIGenerated, processingNotes, sentenceCount)
	if err != nil {
		return fmt.Errorf("update processed content: %w", err)
	}
	return nil
}

func (r *ArticleRepository) UpdateProcessingState(
	ctx context.Context,
	articleID int64,
	translationStatus model.TranslationStatus,
	processingNotes *string,
) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE articles
		SET translation_status = $2,
		    processing_notes = $3,
		    updated_at = NOW()
		WHERE id = $1
	`, articleID, translationStatus, processingNotes)
	if err != nil {
		return fmt.Errorf("update processing state: %w", err)
	}
	return nil
}

func (r *ArticleRepository) ReplaceSentences(ctx context.Context, articleID int64, sentences []model.ArticleSentence) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin article sentence tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM article_sentences WHERE article_id = $1`, articleID); err != nil {
		return fmt.Errorf("delete article sentences: %w", err)
	}

	for _, sentence := range sentences {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO article_sentences (article_id, sentence_order, sentence_text, translation_zh)
			VALUES ($1, $2, $3, $4)
		`, articleID, sentence.SentenceOrder, sentence.SentenceText, sentence.TranslationZH); err != nil {
			return fmt.Errorf("insert article sentence: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit article sentence tx: %w", err)
	}
	return nil
}

func (r *ArticleRepository) ListSentences(ctx context.Context, articleID int64) ([]model.ArticleSentence, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, article_id, sentence_order, sentence_text, translation_zh, created_at
		FROM article_sentences
		WHERE article_id = $1
		ORDER BY sentence_order ASC
	`, articleID)
	if err != nil {
		return nil, fmt.Errorf("list article sentences: %w", err)
	}
	defer rows.Close()

	var sentences []model.ArticleSentence
	for rows.Next() {
		var sentence model.ArticleSentence
		if err := rows.Scan(
			&sentence.ID,
			&sentence.ArticleID,
			&sentence.SentenceOrder,
			&sentence.SentenceText,
			&sentence.TranslationZH,
			&sentence.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan article sentence: %w", err)
		}
		sentences = append(sentences, sentence)
	}
	return sentences, rows.Err()
}

type articleScanner interface {
	Scan(dest ...any) error
}

func scanArticle(scanner articleScanner) (*model.Article, error) {
	var article model.Article
	if err := scanner.Scan(
		&article.ID,
		&article.UserID,
		&article.Title,
		&article.OriginalLanguage,
		&article.OriginalContent,
		&article.JapaneseContent,
		&article.ChineseTranslation,
		&article.JLPTLevel,
		&article.SourceType,
		&article.IsAIGenerated,
		&article.IsVerified,
		&article.TranslationStatus,
		&article.ProcessingNotes,
		&article.SentenceCount,
		&article.CreatedAt,
		&article.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &article, nil
}
