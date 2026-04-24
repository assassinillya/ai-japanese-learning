package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ai-japanese-learning/internal/model"
	"ai-japanese-learning/internal/repository"
)

type VocabularyService struct {
	vocabularyRepo *repository.VocabularyRepository
	dictionaryRepo *repository.DictionaryRepository
	articleRepo    *repository.ArticleRepository
}

func NewVocabularyService(
	vocabularyRepo *repository.VocabularyRepository,
	dictionaryRepo *repository.DictionaryRepository,
	articleRepo *repository.ArticleRepository,
) *VocabularyService {
	return &VocabularyService{
		vocabularyRepo: vocabularyRepo,
		dictionaryRepo: dictionaryRepo,
		articleRepo:    articleRepo,
	}
}

func (s *VocabularyService) Check(ctx context.Context, userID, dictionaryEntryID int64) (*model.UserVocabulary, bool, error) {
	item, err := s.vocabularyRepo.GetByUserAndDictionaryEntry(ctx, userID, dictionaryEntryID)
	if err == nil {
		return item, true, nil
	}
	if err == repository.ErrVocabularyNotFound {
		return nil, false, nil
	}
	return nil, false, err
}

func (s *VocabularyService) Add(
	ctx context.Context,
	userID int64,
	dictionaryEntryID int64,
	articleID *int64,
	sourceSentenceID *int64,
	selectedText string,
	sourceSentenceText string,
) (*model.UserVocabulary, bool, error) {
	selectedText = strings.TrimSpace(selectedText)
	sourceSentenceText = strings.TrimSpace(sourceSentenceText)
	if selectedText == "" || sourceSentenceText == "" {
		return nil, false, fmt.Errorf("selected_text and source_sentence_text are required")
	}

	if _, err := s.dictionaryRepo.GetByID(ctx, dictionaryEntryID); err != nil {
		return nil, false, err
	}
	if articleID != nil {
		if _, err := s.articleRepo.GetAccessible(ctx, userID, *articleID); err != nil {
			return nil, false, err
		}
	}

	existing, err := s.vocabularyRepo.GetByUserAndDictionaryEntry(ctx, userID, dictionaryEntryID)
	if err == nil {
		return existing, false, nil
	}
	if err != repository.ErrVocabularyNotFound {
		return nil, false, err
	}

	item := &model.UserVocabulary{
		UserID:                  userID,
		DictionaryEntryID:       dictionaryEntryID,
		ArticleID:               articleID,
		SourceSentenceID:        sourceSentenceID,
		SelectedText:            selectedText,
		SourceSentenceText:      sourceSentenceText,
		Status:                  model.VocabularyNew,
		Familiarity:             0,
		CorrectCount:            0,
		WrongCount:              0,
		ConsecutiveCorrectCount: 0,
		NextReviewAt:            time.Now(),
	}

	created, err := s.vocabularyRepo.Create(ctx, item)
	if err != nil {
		return nil, false, err
	}
	return created, true, nil
}
