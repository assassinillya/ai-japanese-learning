package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"time"

	"ai-japanese-learning/internal/model"
	"ai-japanese-learning/internal/repository"
)

type ReviewService struct {
	vocabularyRepo *repository.VocabularyRepository
	dictionaryRepo *repository.DictionaryRepository
	reviewRepo     *repository.ReviewRepository
	aiService      *AIService
}

func NewReviewService(
	vocabularyRepo *repository.VocabularyRepository,
	dictionaryRepo *repository.DictionaryRepository,
	reviewRepo *repository.ReviewRepository,
	aiService *AIService,
) *ReviewService {
	return &ReviewService{
		vocabularyRepo: vocabularyRepo,
		dictionaryRepo: dictionaryRepo,
		reviewRepo:     reviewRepo,
		aiService:      aiService,
	}
}

func (s *ReviewService) Due(ctx context.Context, userID int64, limit int) ([]model.VocabularyReviewItem, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	dueItems, err := s.vocabularyRepo.ListDueForReview(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	items := make([]model.VocabularyReviewItem, 0, len(dueItems))
	for _, detail := range dueItems {
		question, err := s.GetOrCreateQuestion(ctx, detail.DictionaryEntry)
		if err != nil {
			return nil, err
		}
		items = append(items, model.VocabularyReviewItem{
			UserVocabulary:  detail.Item,
			Dictionary:      detail.DictionaryEntry,
			Question:        *question,
			ArticleTitle:    detail.ArticleTitle,
			ContextSentence: detail.ExampleSentence,
		})
	}
	return items, nil
}

func (s *ReviewService) GetOrCreateQuestion(ctx context.Context, entry model.DictionaryEntry) (*model.VocabularyReviewQuestion, error) {
	existing, err := s.reviewRepo.GetQuestionByDictionaryEntry(ctx, entry.ID)
	if err == nil {
		return existing, nil
	}
	if err != repository.ErrReviewQuestionNotFound {
		return nil, err
	}

	fallbackModel := "placeholder-vocabulary-review-generator"
	aiModel := fallbackModel
	if s.aiService != nil {
		aiModel = s.aiService.ModelName(fallbackModel)
	}
	promptVersion := aiPromptVersionV12
	taskType := "vocabulary_review_question"
	prompt := promptReviewQuestion(entry)
	request := reviewQuestionCacheRequest{
		DictionaryEntryID: entry.ID,
		Surface:           entry.Surface,
		PrimaryMeaningZH:  entry.PrimaryMeaningZH,
		MeaningZH:         entry.MeaningZH,
		PartOfSpeech:      entry.PartOfSpeech,
		JLPTLevel:         entry.JLPTLevel,
		Prompt:            prompt,
	}
	cacheKey, inputHash := s.reviewQuestionCacheKey(request, taskType, aiModel, promptVersion)
	if cached, ok := s.getCachedReviewQuestion(ctx, cacheKey); ok {
		cached.DictionaryEntryID = entry.ID
		cached.ID = 0
		cached.CreatedAt = time.Time{}
		return s.reviewRepo.CreateQuestion(ctx, cached)
	}

	question, err := s.generateReviewQuestionWithAI(ctx, entry, prompt)
	if err != nil {
		if s.aiService != nil {
			s.aiService.LogFailure(ctx, taskType, request, err, aiModel, promptVersion)
		}
		options, correctOption, err := s.buildMeaningOptions(ctx, entry)
		if err != nil {
			return nil, err
		}
		question = &model.VocabularyReviewQuestion{
			DictionaryEntryID: entry.ID,
			QuestionText:      entry.Surface,
			CorrectAnswer:     meaningForReview(entry),
			OptionA:           options[0],
			OptionB:           options[1],
			OptionC:           options[2],
			OptionD:           options[3],
			CorrectOption:     correctOption,
			ExplanationZH:     fmt.Sprintf("「%s」的主要中文意思是：%s。", entry.Surface, meaningForReview(entry)),
		}
	}
	question.AIModel = &aiModel
	question.PromptVersion = &promptVersion
	s.storeCachedReviewQuestion(ctx, taskType, inputHash, cacheKey, request, question, aiModel, promptVersion)
	return s.reviewRepo.CreateQuestion(ctx, question)
}

func (s *ReviewService) Records(ctx context.Context, userID int64, limit int) ([]model.VocabularyReviewRecordDetail, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.reviewRepo.ListRecordsByUser(ctx, userID, limit)
}

func (s *ReviewService) SubmitAnswer(ctx context.Context, userID, userVocabularyID, reviewQuestionID int64, selectedOption string) (map[string]any, error) {
	selectedOption = strings.ToUpper(strings.TrimSpace(selectedOption))
	if !slices.Contains([]string{"A", "B", "C", "D"}, selectedOption) {
		return nil, fmt.Errorf("invalid selected option")
	}

	detail, err := s.vocabularyRepo.GetDetail(ctx, userID, userVocabularyID)
	if err != nil {
		return nil, err
	}
	question, err := s.reviewRepo.GetQuestionByID(ctx, reviewQuestionID)
	if err != nil {
		return nil, err
	}
	if question.DictionaryEntryID != detail.DictionaryEntry.ID {
		return nil, fmt.Errorf("review question does not match vocabulary item")
	}

	isCorrect := question.CorrectOption == selectedOption
	record := &model.VocabularyReviewRecord{
		UserID:           userID,
		UserVocabularyID: userVocabularyID,
		ReviewQuestionID: reviewQuestionID,
		SelectedOption:   selectedOption,
		IsCorrect:        isCorrect,
	}
	if _, err := s.reviewRepo.CreateRecord(ctx, record); err != nil {
		return nil, err
	}

	status, familiarity, correctCount, wrongCount, consecutive, nextReviewAt := nextReviewState(detail.Item, isCorrect)
	if err := s.vocabularyRepo.UpdateReviewProgress(
		ctx,
		userID,
		userVocabularyID,
		status,
		familiarity,
		correctCount,
		wrongCount,
		consecutive,
		nextReviewAt,
	); err != nil {
		return nil, err
	}

	return map[string]any{
		"is_correct":                isCorrect,
		"correct_option":            question.CorrectOption,
		"correct_answer":            question.CorrectAnswer,
		"explanation":               question.ExplanationZH,
		"next_review_at":            nextReviewAt,
		"status":                    status,
		"familiarity":               familiarity,
		"correct_count":             correctCount,
		"wrong_count":               wrongCount,
		"consecutive_correct_count": consecutive,
	}, nil
}

func (s *ReviewService) buildMeaningOptions(ctx context.Context, entry model.DictionaryEntry) ([]string, string, error) {
	correctMeaning := meaningForReview(entry)
	optionSet := []string{correctMeaning}

	distractors, err := s.dictionaryRepo.ListDistractors(ctx, entry.ID, entry.PartOfSpeech, 10)
	if err != nil {
		return nil, "", err
	}
	for _, distractor := range distractors {
		meaning := meaningForReview(distractor)
		if meaning != "" && !slices.Contains(optionSet, meaning) {
			optionSet = append(optionSet, meaning)
		}
		if len(optionSet) == 4 {
			break
		}
	}
	for _, fallback := range []string{"取消", "购买", "出发", "查看", "听见", "书写"} {
		if len(optionSet) == 4 {
			break
		}
		if !slices.Contains(optionSet, fallback) {
			optionSet = append(optionSet, fallback)
		}
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(entry.ID)*17))
	rng.Shuffle(len(optionSet), func(i, j int) {
		optionSet[i], optionSet[j] = optionSet[j], optionSet[i]
	})

	correctOption := "A"
	for idx, option := range optionSet {
		if option == correctMeaning {
			correctOption = string(rune('A' + idx))
			break
		}
	}
	return optionSet, correctOption, nil
}

func meaningForReview(entry model.DictionaryEntry) string {
	if strings.TrimSpace(entry.PrimaryMeaningZH) != "" {
		return strings.TrimSpace(entry.PrimaryMeaningZH)
	}
	return strings.TrimSpace(entry.MeaningZH)
}

type reviewQuestionCacheRequest struct {
	DictionaryEntryID int64    `json:"dictionary_entry_id"`
	Surface           string   `json:"surface"`
	PrimaryMeaningZH  string   `json:"primary_meaning_zh"`
	MeaningZH         string   `json:"meaning_zh"`
	PartOfSpeech      string   `json:"part_of_speech"`
	JLPTLevel         string   `json:"jlpt_level"`
	Prompt            AIPrompt `json:"prompt"`
}

func (s *ReviewService) generateReviewQuestionWithAI(ctx context.Context, entry model.DictionaryEntry, prompt AIPrompt) (*model.VocabularyReviewQuestion, error) {
	if s.aiService == nil || !s.aiService.ProviderAvailable() {
		return nil, fmt.Errorf("ai provider unavailable")
	}
	raw, err := s.aiService.CompleteJSON(ctx, prompt)
	if err != nil {
		return nil, err
	}
	var question model.VocabularyReviewQuestion
	if err := json.Unmarshal([]byte(raw), &question); err != nil {
		return nil, fmt.Errorf("parse ai review question: %w", err)
	}
	question.DictionaryEntryID = entry.ID
	if strings.TrimSpace(question.QuestionText) == "" {
		question.QuestionText = entry.Surface
	}
	if question.CorrectAnswer != meaningForReview(entry) {
		return nil, fmt.Errorf("ai review question correct_answer does not match dictionary primary meaning")
	}
	if !slices.Contains([]string{"A", "B", "C", "D"}, question.CorrectOption) {
		return nil, fmt.Errorf("ai review question invalid correct option")
	}
	if strings.TrimSpace(question.OptionA) == "" || strings.TrimSpace(question.OptionB) == "" ||
		strings.TrimSpace(question.OptionC) == "" || strings.TrimSpace(question.OptionD) == "" {
		return nil, fmt.Errorf("ai review question missing options")
	}
	if strings.TrimSpace(question.ExplanationZH) == "" {
		question.ExplanationZH = fmt.Sprintf("「%s」的主要中文意思是：%s。", entry.Surface, meaningForReview(entry))
	}
	return &question, nil
}

func (s *ReviewService) reviewQuestionCacheKey(request reviewQuestionCacheRequest, taskType, aiModel, promptVersion string) (string, string) {
	if s.aiService == nil {
		return "", ""
	}
	raw, _ := json.Marshal(request)
	inputHash := s.aiService.HashInput(string(raw))
	return s.aiService.CacheKey(taskType, inputHash, aiModel, promptVersion), inputHash
}

func (s *ReviewService) getCachedReviewQuestion(ctx context.Context, cacheKey string) (*model.VocabularyReviewQuestion, bool) {
	if s.aiService == nil || cacheKey == "" {
		return nil, false
	}
	cached, ok, err := s.aiService.GetCached(ctx, cacheKey)
	if err != nil || !ok {
		return nil, false
	}
	var question model.VocabularyReviewQuestion
	if err := json.Unmarshal([]byte(cached.ResponseJSON), &question); err != nil || strings.TrimSpace(question.QuestionText) == "" {
		return nil, false
	}
	return &question, true
}

func (s *ReviewService) storeCachedReviewQuestion(ctx context.Context, taskType, inputHash, cacheKey string, request reviewQuestionCacheRequest, question *model.VocabularyReviewQuestion, aiModel, promptVersion string) {
	if s.aiService == nil || cacheKey == "" {
		return
	}
	cacheQuestion := *question
	cacheQuestion.ID = 0
	cacheQuestion.CreatedAt = time.Time{}
	if _, err := s.aiService.StoreCached(ctx, taskType, inputHash, cacheKey, request, cacheQuestion, aiModel, promptVersion); err != nil {
		s.aiService.LogFailure(ctx, taskType, request, err, aiModel, promptVersion)
	}
}

func nextReviewState(item model.UserVocabulary, isCorrect bool) (model.VocabularyStatus, int, int, int, int, time.Time) {
	now := time.Now()
	familiarity := item.Familiarity
	correctCount := item.CorrectCount
	wrongCount := item.WrongCount
	consecutive := item.ConsecutiveCorrectCount

	if !isCorrect {
		wrongCount++
		consecutive = 0
		if familiarity > 0 {
			familiarity--
		}
		return model.VocabularyLearning, familiarity, correctCount, wrongCount, consecutive, now.Add(10 * time.Minute)
	}

	correctCount++
	consecutive++
	if familiarity < 5 {
		familiarity++
	}

	switch {
	case consecutive >= 5:
		return model.VocabularyMastered, familiarity, correctCount, wrongCount, consecutive, now.Add(30 * 24 * time.Hour)
	case consecutive == 4:
		return model.VocabularyReviewing, familiarity, correctCount, wrongCount, consecutive, now.Add(15 * 24 * time.Hour)
	case consecutive == 3:
		return model.VocabularyReviewing, familiarity, correctCount, wrongCount, consecutive, now.Add(7 * 24 * time.Hour)
	case consecutive == 2:
		return model.VocabularyLearning, familiarity, correctCount, wrongCount, consecutive, now.Add(3 * 24 * time.Hour)
	default:
		return model.VocabularyLearning, familiarity, correctCount, wrongCount, consecutive, now.Add(24 * time.Hour)
	}
}
