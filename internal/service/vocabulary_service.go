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

func isValidVocabularyStatus(status model.VocabularyStatus) bool {
	switch status {
	case model.VocabularyNew, model.VocabularyLearning, model.VocabularyReviewing, model.VocabularyMastered, model.VocabularyIgnored:
		return true
	default:
		return false
	}
}

func sameOptionalInt64(left, right *int64) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
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

func (s *VocabularyService) List(ctx context.Context, userID int64, status string, search string) ([]model.VocabularyDetail, error) {
	if status != "" && !isValidVocabularyStatus(model.VocabularyStatus(status)) {
		return nil, fmt.Errorf("invalid vocabulary status")
	}
	return s.vocabularyRepo.ListByUser(ctx, userID, status, search)
}

func (s *VocabularyService) GetDetail(ctx context.Context, userID, vocabularyID int64) (*model.VocabularyDetail, error) {
	return s.vocabularyRepo.GetDetail(ctx, userID, vocabularyID)
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
		if existing.SourceSentenceText != sourceSentenceText ||
			existing.SelectedText != selectedText ||
			!sameOptionalInt64(existing.SourceSentenceID, sourceSentenceID) ||
			!sameOptionalInt64(existing.ArticleID, articleID) {
			if err := s.vocabularyRepo.UpdateContext(ctx, userID, dictionaryEntryID, articleID, sourceSentenceID, selectedText, sourceSentenceText); err != nil {
				return nil, false, err
			}
			updated, err := s.vocabularyRepo.GetByUserAndDictionaryEntry(ctx, userID, dictionaryEntryID)
			return updated, false, err
		}
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

func (s *VocabularyService) UpdateStatus(ctx context.Context, userID, vocabularyID int64, status model.VocabularyStatus) (*model.VocabularyDetail, error) {
	if !isValidVocabularyStatus(status) {
		return nil, fmt.Errorf("invalid vocabulary status")
	}
	if err := s.vocabularyRepo.UpdateStatus(ctx, userID, vocabularyID, status); err != nil {
		return nil, err
	}
	return s.vocabularyRepo.GetDetail(ctx, userID, vocabularyID)
}

func (s *VocabularyService) UpdateStatusBatch(ctx context.Context, userID int64, vocabularyIDs []int64, status model.VocabularyStatus) (int64, error) {
	if !isValidVocabularyStatus(status) {
		return 0, fmt.Errorf("invalid vocabulary status")
	}
	vocabularyIDs = normalizeVocabularyIDs(vocabularyIDs)
	if len(vocabularyIDs) == 0 {
		return 0, fmt.Errorf("vocabulary_ids are required")
	}
	return s.vocabularyRepo.UpdateStatusBatch(ctx, userID, vocabularyIDs, status)
}

func (s *VocabularyService) Delete(ctx context.Context, userID, vocabularyID int64) error {
	return s.vocabularyRepo.Delete(ctx, userID, vocabularyID)
}

func (s *VocabularyService) DeleteBatch(ctx context.Context, userID int64, vocabularyIDs []int64) (int64, error) {
	vocabularyIDs = normalizeVocabularyIDs(vocabularyIDs)
	if len(vocabularyIDs) == 0 {
		return 0, fmt.Errorf("vocabulary_ids are required")
	}
	return s.vocabularyRepo.DeleteBatch(ctx, userID, vocabularyIDs)
}

func normalizeVocabularyIDs(ids []int64) []int64 {
	seen := map[int64]bool{}
	normalized := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		normalized = append(normalized, id)
	}
	return normalized
}
