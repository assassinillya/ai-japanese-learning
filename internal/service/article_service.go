package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"ai-japanese-learning/internal/model"
	"ai-japanese-learning/internal/repository"
)

type ArticleService struct {
	articleRepo        *repository.ArticleRepository
	languageService    *LanguageService
	translationService *TranslationService
}

const (
	maxArticleTitleLength   = 120
	maxArticleContentLength = 12000
)

func NewArticleService(
	articleRepo *repository.ArticleRepository,
	languageService *LanguageService,
	translationService *TranslationService,
) *ArticleService {
	return &ArticleService{
		articleRepo:        articleRepo,
		languageService:    languageService,
		translationService: translationService,
	}
}

func (s *ArticleService) ListLibrary(ctx context.Context) ([]model.Article, error) {
	return s.articleRepo.ListLibrary(ctx)
}

func (s *ArticleService) ListByUser(ctx context.Context, userID int64) ([]model.Article, error) {
	return s.articleRepo.ListByUser(ctx, userID)
}

func (s *ArticleService) Create(ctx context.Context, userID int64, title, content string, level model.JLPTLevel) (*model.Article, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if title == "" || content == "" {
		return nil, fmt.Errorf("title and content are required")
	}
	if utf8.RuneCountInString(title) > maxArticleTitleLength {
		return nil, fmt.Errorf("title is too long")
	}
	if utf8.RuneCountInString(content) > maxArticleContentLength {
		return nil, fmt.Errorf("article content is too long")
	}

	language := s.languageService.Detect(content)
	originalContent := content
	note := "Article uploaded successfully and is ready for processing."

	article := &model.Article{
		UserID:            &userID,
		Title:             title,
		OriginalLanguage:  language,
		OriginalContent:   &originalContent,
		JapaneseContent:   "",
		JLPTLevel:         level,
		SourceType:        "user_uploaded",
		IsAIGenerated:     false,
		IsVerified:        false,
		TranslationStatus: model.TranslationPending,
		ProcessingNotes:   &note,
		SentenceCount:     0,
	}

	created, err := s.articleRepo.Create(ctx, article)
	if err != nil {
		return nil, err
	}

	return s.Process(ctx, userID, created.ID)
}

func (s *ArticleService) Get(ctx context.Context, userID, articleID int64) (*model.Article, error) {
	return s.articleRepo.GetAccessible(ctx, userID, articleID)
}

func (s *ArticleService) Process(ctx context.Context, userID, articleID int64) (*model.Article, error) {
	article, err := s.articleRepo.GetAccessible(ctx, userID, articleID)
	if err != nil {
		return nil, err
	}
	if article.UserID == nil || *article.UserID != userID {
		return nil, errors.New("article is not owned by current user")
	}
	if article.OriginalContent == nil || strings.TrimSpace(*article.OriginalContent) == "" {
		return nil, fmt.Errorf("article content is empty")
	}

	processingNote := "Processing article content and splitting sentences."
	if err := s.articleRepo.UpdateProcessingState(ctx, article.ID, model.TranslationProcessing, &processingNote); err != nil {
		return nil, err
	}

	translated := strings.TrimSpace(article.JapaneseContent)
	sourceType := article.SourceType
	aiGenerated := article.IsAIGenerated
	note := "Rebuilt sentences from stored Japanese article content."
	if translated == "" {
		translated, sourceType, aiGenerated, note = s.translationService.TranslateToJapanese(
			ctx,
			article.OriginalLanguage,
			*article.OriginalContent,
			article.JLPTLevel,
		)
	}
	if translated == "" {
		failedNote := "Article processing failed because the translated content is empty."
		_ = s.articleRepo.UpdateProcessingState(ctx, article.ID, model.TranslationFailed, &failedNote)
		return nil, fmt.Errorf("translated article content is empty")
	}

	rawSentences := splitSentences(translated)
	sentences := make([]model.ArticleSentence, 0, len(rawSentences))
	for idx, sentence := range rawSentences {
		sentences = append(sentences, model.ArticleSentence{
			ArticleID:     article.ID,
			SentenceOrder: idx + 1,
			SentenceText:  sentence,
			CreatedAt:     time.Now(),
		})
	}

	if err := s.articleRepo.ReplaceSentences(ctx, article.ID, sentences); err != nil {
		failedNote := "Article processing failed while saving split sentences."
		_ = s.articleRepo.UpdateProcessingState(ctx, article.ID, model.TranslationFailed, &failedNote)
		return nil, err
	}
	if err := s.articleRepo.UpdateProcessedContent(
		ctx,
		article.ID,
		translated,
		model.TranslationDone,
		sourceType,
		aiGenerated,
		&note,
		len(sentences),
	); err != nil {
		failedNote := "Article processing failed while saving processed article content."
		_ = s.articleRepo.UpdateProcessingState(ctx, article.ID, model.TranslationFailed, &failedNote)
		return nil, err
	}

	return s.articleRepo.GetAccessible(ctx, userID, articleID)
}

func (s *ArticleService) ListSentences(ctx context.Context, userID, articleID int64) ([]model.ArticleSentence, error) {
	if _, err := s.articleRepo.GetAccessible(ctx, userID, articleID); err != nil {
		return nil, err
	}
	return s.articleRepo.ListSentences(ctx, articleID)
}

func splitSentences(content string) []string {
	cleaned := strings.TrimSpace(content)
	if cleaned == "" {
		return nil
	}

	cleaned = strings.ReplaceAll(cleaned, "\r\n", "\n")
	cleaned = strings.ReplaceAll(cleaned, "\n", " ")
	cleaned = strings.NewReplacer("。", "。\n", "！", "！\n", "？", "？\n").Replace(cleaned)

	parts := strings.Split(cleaned, "\n")
	var sentences []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			sentences = append(sentences, part)
		}
	}
	if len(sentences) == 0 {
		return []string{strings.TrimSpace(content)}
	}
	return sentences
}
