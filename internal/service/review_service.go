package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	return s.reviewItems(ctx, userID, limit, false)
}

func (s *ReviewService) Extra(ctx context.Context, userID int64, limit int) ([]model.VocabularyReviewItem, error) {
	return s.reviewItems(ctx, userID, limit, true)
}

func (s *ReviewService) reviewItems(ctx context.Context, userID int64, limit int, includeFuture bool) ([]model.VocabularyReviewItem, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var dueItems []model.VocabularyDetail
	var err error
	if includeFuture {
		dueItems, err = s.vocabularyRepo.ListExtraForReview(ctx, userID, limit)
	} else {
		dueItems, err = s.vocabularyRepo.ListDueForReview(ctx, userID, limit)
	}
	if err != nil {
		return nil, err
	}

	items := make([]model.VocabularyReviewItem, 0, len(dueItems))
	for _, detail := range dueItems {
		question, err := s.NextQuestion(ctx, userID, detail.DictionaryEntry)
		if err != nil {
			return nil, err
		}
		dailyStats, err := s.reviewRepo.DailyStats(ctx, userID, detail.Item.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, model.VocabularyReviewItem{
			UserVocabulary:    detail.Item,
			Dictionary:        detail.DictionaryEntry,
			Question:          *question,
			ArticleTitle:      detail.ArticleTitle,
			ContextSentence:   detail.ExampleSentence,
			TodayReviewCount:  dailyStats.ReviewCount,
			TodayCorrectCount: dailyStats.CorrectCount,
			TodayWrongCount:   dailyStats.WrongCount,
			TodayProgressGain: dailyStats.FamiliarityGain,
		})
	}
	return items, nil
}

func (s *ReviewService) EnsureQuestionsForUser(ctx context.Context, userID int64) (int, error) {
	details, err := s.vocabularyRepo.ListByUser(ctx, userID, "", "")
	if err != nil {
		return 0, err
	}
	created := 0
	for _, detail := range details {
		if detail.Item.Status == model.VocabularyIgnored || detail.Item.Status == model.VocabularyMastered {
			continue
		}
		count, err := s.reviewRepo.CountQuestionsByDictionaryEntry(ctx, detail.DictionaryEntry.ID)
		if err == nil && count >= 3 {
			continue
		} else if err != nil {
			return created, err
		}
		if _, err := s.EnsureQuestionsForEntry(ctx, detail.DictionaryEntry); err != nil {
			return created, err
		}
		created += max(0, 3-count)
	}
	return created, nil
}

func (s *ReviewService) EnsureQuestionsForAllVocabulary(ctx context.Context) (int, error) {
	entries, err := s.vocabularyRepo.ListDictionaryEntriesMissingReviewQuestions(ctx, 3)
	if err != nil {
		return 0, err
	}
	created := 0
	for _, entry := range entries {
		before, err := s.reviewRepo.CountQuestionsByDictionaryEntry(ctx, entry.ID)
		if err != nil {
			return created, err
		}
		if _, err := s.EnsureQuestionsForEntry(ctx, entry); err != nil {
			return created, err
		}
		after, err := s.reviewRepo.CountQuestionsByDictionaryEntry(ctx, entry.ID)
		if err != nil {
			return created, err
		}
		created += max(0, after-before)
	}
	return created, nil
}

func (s *ReviewService) PrewarmMissingQuestionsAsync(ctx context.Context) {
	go func() {
		created, err := s.EnsureQuestionsForAllVocabulary(ctx)
		if err != nil {
			log.Printf("prewarm vocabulary review questions failed: %v", err)
			return
		}
		if created > 0 {
			log.Printf("prewarmed %d vocabulary review questions", created)
		}
	}()
}

func (s *ReviewService) NextQuestion(ctx context.Context, userID int64, entry model.DictionaryEntry) (*model.VocabularyReviewQuestion, error) {
	if count, err := s.reviewRepo.CountQuestionsByDictionaryEntry(ctx, entry.ID); err == nil && count < 3 {
		go func() {
			_, _ = s.EnsureQuestionsForEntry(context.Background(), entry)
		}()
	}
	question, err := s.reviewRepo.NextQuestionForUser(ctx, userID, entry.ID)
	if err == nil {
		return question, nil
	}
	if err != repository.ErrReviewQuestionNotFound {
		return nil, err
	}
	return s.GetOrCreateQuestion(ctx, entry)
}

func (s *ReviewService) GetOrCreateQuestion(ctx context.Context, entry model.DictionaryEntry) (*model.VocabularyReviewQuestion, error) {
	return s.getOrCreateQuestion(ctx, entry, 1)
}

func (s *ReviewService) EnsureQuestionsForEntry(ctx context.Context, entry model.DictionaryEntry) ([]model.VocabularyReviewQuestion, error) {
	questions := make([]model.VocabularyReviewQuestion, 0, 3)
	for order := 1; order <= 3; order++ {
		question, err := s.getOrCreateQuestion(ctx, entry, order)
		if err != nil {
			return nil, err
		}
		questions = append(questions, *question)
	}
	return questions, nil
}

func (s *ReviewService) getOrCreateQuestion(ctx context.Context, entry model.DictionaryEntry, order int) (*model.VocabularyReviewQuestion, error) {
	existing, err := s.reviewRepo.GetQuestionByDictionaryEntryAndOrder(ctx, entry.ID, order)
	if err == nil {
		return existing, nil
	}
	if err != nil && err != repository.ErrReviewQuestionNotFound {
		return nil, err
	}

	fallbackModel := "placeholder-vocabulary-review-generator"
	aiModel := fallbackModel
	if s.aiService != nil {
		aiModel = s.aiService.ModelNameFor(ctx, fallbackModel)
	}
	promptVersion := aiPromptVersionV12
	taskType := "vocabulary_review_question"
	prompt := promptReviewQuestion(entry, order)
	request := reviewQuestionCacheRequest{
		DictionaryEntryID: entry.ID,
		QuestionOrder:     order,
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
		cached.QuestionOrder = order
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
			QuestionOrder:     order,
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
	question.QuestionOrder = order
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
	dailyStats, err := s.reviewRepo.DailyStats(ctx, userID, userVocabularyID)
	if err != nil {
		return nil, err
	}

	status, familiarity, correctCount, wrongCount, consecutive, nextReviewAt, familiarityDelta := nextReviewState(detail.Item, isCorrect, dailyStats)
	record := &model.VocabularyReviewRecord{
		UserID:           userID,
		UserVocabularyID: userVocabularyID,
		ReviewQuestionID: reviewQuestionID,
		SelectedOption:   selectedOption,
		IsCorrect:        isCorrect,
		FamiliarityDelta: familiarityDelta,
	}
	if _, err := s.reviewRepo.CreateRecord(ctx, record); err != nil {
		return nil, err
	}

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
		"proficiency":               familiarity,
		"familiarity_delta":         familiarityDelta,
		"today_progress_gain":       dailyStats.FamiliarityGain + familiarityDelta,
		"daily_gain_cap_reached":    dailyStats.FamiliarityGain+familiarityDelta >= 40,
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
	QuestionOrder     int      `json:"question_order"`
	Surface           string   `json:"surface"`
	PrimaryMeaningZH  string   `json:"primary_meaning_zh"`
	MeaningZH         string   `json:"meaning_zh"`
	PartOfSpeech      string   `json:"part_of_speech"`
	JLPTLevel         string   `json:"jlpt_level"`
	Prompt            AIPrompt `json:"prompt"`
}

func (s *ReviewService) generateReviewQuestionWithAI(ctx context.Context, entry model.DictionaryEntry, prompt AIPrompt) (*model.VocabularyReviewQuestion, error) {
	if s.aiService == nil || !s.aiService.ProviderAvailableFor(ctx) {
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

func nextReviewState(item model.UserVocabulary, isCorrect bool, dailyStats repository.DailyReviewStats) (model.VocabularyStatus, int, int, int, int, time.Time, int) {
	now := time.Now()
	familiarity := normalizeProficiency(item.Familiarity)
	correctCount := item.CorrectCount
	wrongCount := item.WrongCount
	consecutive := item.ConsecutiveCorrectCount

	if !isCorrect {
		wrongCount++
		consecutive = 0
		return model.VocabularyLearning, familiarity, correctCount, wrongCount, consecutive, now.Add(nextReviewDuration(familiarity, dailyStats.CorrectCount, dailyStats.WrongCount+1)), 0
	}

	correctCount++
	consecutive++
	correctOrdinal := dailyStats.CorrectCount + 1
	adjustedCap := max(0, 40-dailyStats.FamiliarityGain)
	previousGain := min(theoreticalCorrectGain(dailyStats.CorrectCount), adjustedCap)
	remainingGain := max(0, adjustedCap-previousGain)
	gain := min(correctGain(correctOrdinal), remainingGain)
	familiarity = min(100, familiarity+gain)

	if familiarity >= 100 {
		return model.VocabularyMastered, 100, correctCount, wrongCount, consecutive, now.Add(3650 * 24 * time.Hour), gain
	}
	return model.VocabularyLearning, familiarity, correctCount, wrongCount, consecutive, now.Add(nextReviewDuration(familiarity, dailyStats.CorrectCount+1, dailyStats.WrongCount)), gain
}

func normalizeProficiency(value int) int {
	if value < 0 {
		return 0
	}
	if value <= 5 {
		return value * 20
	}
	if value > 100 {
		return 100
	}
	return value
}

func dailyGainCap(proficiency int) int {
	switch {
	case proficiency <= 20:
		return 25
	case proficiency <= 50:
		return 20
	case proficiency <= 80:
		return 15
	default:
		return 10
	}
}

func adjustedDailyGainCap(baseCap, wrongToday int) int {
	rate := 1.0
	switch {
	case wrongToday == 1:
		rate = 0.7
	case wrongToday == 2:
		rate = 0.5
	case wrongToday >= 3:
		rate = 0.35
	}
	return max(5, int(float64(baseCap)*rate+0.5))
}

func correctGain(correctOrdinal int) int {
	switch correctOrdinal {
	case 1:
		return 8
	case 2:
		return 6
	case 3:
		return 4
	default:
		return 2
	}
}

func theoreticalCorrectGain(correctCount int) int {
	total := 0
	for i := 1; i <= correctCount; i++ {
		total += correctGain(i)
	}
	return total
}

func nextReviewDuration(proficiency, correctToday, wrongToday int) time.Duration {
	baseDays := 1
	switch {
	case proficiency <= 20:
		baseDays = 1
	case proficiency <= 50:
		baseDays = 2
	case proficiency <= 80:
		baseDays = 4
	default:
		baseDays = 7
	}
	total := correctToday + wrongToday
	rate := 1.0
	if total > 0 {
		accuracy := float64(correctToday) / float64(total)
		switch {
		case accuracy >= 1:
			rate = 1.3
		case accuracy >= 0.8:
			rate = 1.0
		case accuracy >= 0.5:
			rate = 0.7
		default:
			rate = 0.5
		}
	}
	days := int(float64(baseDays)*rate + 0.5)
	days = max(1, min(7, days))
	return time.Duration(days) * 24 * time.Hour
}
