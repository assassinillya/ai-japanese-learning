package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"time"
	"unicode"

	"ai-japanese-learning/internal/model"
	"ai-japanese-learning/internal/repository"
)

type ChallengeService struct {
	articleRepo    *repository.ArticleRepository
	dictionaryRepo *repository.DictionaryRepository
	challengeRepo  *repository.ChallengeRepository
	dictionarySvc  *DictionaryService
	aiService      *AIService
}

const (
	QuestionTypeChallengeReading = "challenge_reading"
	QuestionTypePostReadingQuiz  = "post_reading_quiz"
)

func NewChallengeService(
	articleRepo *repository.ArticleRepository,
	dictionaryRepo *repository.DictionaryRepository,
	challengeRepo *repository.ChallengeRepository,
	dictionarySvc *DictionaryService,
	aiService *AIService,
) *ChallengeService {
	return &ChallengeService{
		articleRepo:    articleRepo,
		dictionaryRepo: dictionaryRepo,
		challengeRepo:  challengeRepo,
		dictionarySvc:  dictionarySvc,
		aiService:      aiService,
	}
}

func (s *ChallengeService) GetOrGenerate(ctx context.Context, userID, articleID int64) ([]model.ChallengeQuestion, error) {
	if _, err := s.articleRepo.GetAccessible(ctx, userID, articleID); err != nil {
		return nil, err
	}

	existing, err := s.challengeRepo.ListByArticleAndType(ctx, articleID, QuestionTypeChallengeReading)
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		if !containsPlaceholderQuestions(existing) {
			return existing, nil
		}
	}

	return s.Generate(ctx, userID, articleID)
}

func (s *ChallengeService) Generate(ctx context.Context, userID, articleID int64) ([]model.ChallengeQuestion, error) {
	article, err := s.articleRepo.GetAccessible(ctx, userID, articleID)
	if err != nil {
		return nil, err
	}

	sentences, err := s.articleRepo.ListSentences(ctx, articleID)
	if err != nil {
		return nil, err
	}
	if len(sentences) == 0 {
		return nil, fmt.Errorf("article has no sentences")
	}

	aiModel := "placeholder-challenge-generator"
	promptVersion := "v0.8-review"
	request := challengeCacheRequest(article, sentences, QuestionTypeChallengeReading)
	if s.aiService == nil || !s.aiService.ProviderAvailableFor(ctx) {
		return nil, fmt.Errorf("挑战阅读题需要先配置可用 AI Provider，请在个人中心查看 AI 接入说明")
	}
	aiModel = s.aiService.ModelNameFor(ctx, "ai-challenge-generator")
	promptVersion = aiPromptVersionV12
	cacheKey, inputHash := s.challengeCacheKey(request, aiModel, promptVersion)
	if cached, ok := s.getCachedChallengeQuestions(ctx, cacheKey); ok {
		if err := s.challengeRepo.ReplaceByArticleAndType(ctx, articleID, QuestionTypeChallengeReading, cached); err != nil {
			return nil, err
		}
		return s.challengeRepo.ListByArticleAndType(ctx, articleID, QuestionTypeChallengeReading)
	}
	questions, err := s.generateQuestionsWithAI(ctx, articleID, request, QuestionTypeChallengeReading, aiModel, promptVersion)
	if err != nil {
		s.aiService.LogFailure(ctx, QuestionTypeChallengeReading, request, err, aiModel, promptVersion)
		return nil, err
	}
	s.storeCachedChallengeQuestions(ctx, QuestionTypeChallengeReading, inputHash, cacheKey, request, questions, aiModel, promptVersion)
	if err := s.challengeRepo.ReplaceByArticleAndType(ctx, articleID, QuestionTypeChallengeReading, questions); err != nil {
		return nil, err
	}
	return s.challengeRepo.ListByArticleAndType(ctx, articleID, QuestionTypeChallengeReading)
}

func (s *ChallengeService) GetOrGeneratePostQuiz(ctx context.Context, userID, articleID int64) ([]model.ChallengeQuestion, error) {
	if _, err := s.articleRepo.GetAccessible(ctx, userID, articleID); err != nil {
		return nil, err
	}

	existing, err := s.challengeRepo.ListByArticleAndType(ctx, articleID, QuestionTypePostReadingQuiz)
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		if !containsPlaceholderQuestions(existing) {
			return existing, nil
		}
	}

	return s.GeneratePostQuiz(ctx, userID, articleID)
}

func (s *ChallengeService) GeneratePostQuiz(ctx context.Context, userID, articleID int64) ([]model.ChallengeQuestion, error) {
	article, err := s.articleRepo.GetAccessible(ctx, userID, articleID)
	if err != nil {
		return nil, err
	}

	sentences, err := s.articleRepo.ListSentences(ctx, articleID)
	if err != nil {
		return nil, err
	}
	if len(sentences) == 0 {
		return nil, fmt.Errorf("article has no sentences")
	}

	aiModel := "placeholder-post-quiz-generator"
	promptVersion := "v0.8-review"
	request := challengeCacheRequest(article, sentences, QuestionTypePostReadingQuiz)
	if s.aiService == nil || !s.aiService.ProviderAvailableFor(ctx) {
		return nil, fmt.Errorf("阅读后测验需要先配置可用 AI Provider，请在个人中心查看 AI 接入说明")
	}
	aiModel = s.aiService.ModelNameFor(ctx, "ai-post-quiz-generator")
	promptVersion = aiPromptVersionV12
	cacheKey, inputHash := s.challengeCacheKey(request, aiModel, promptVersion)
	if cached, ok := s.getCachedChallengeQuestions(ctx, cacheKey); ok {
		if err := s.challengeRepo.ReplaceByArticleAndType(ctx, articleID, QuestionTypePostReadingQuiz, cached); err != nil {
			return nil, err
		}
		return s.challengeRepo.ListByArticleAndType(ctx, articleID, QuestionTypePostReadingQuiz)
	}
	questions, err := s.generateQuestionsWithAI(ctx, articleID, request, QuestionTypePostReadingQuiz, aiModel, promptVersion)
	if err != nil {
		s.aiService.LogFailure(ctx, QuestionTypePostReadingQuiz, request, err, aiModel, promptVersion)
		return nil, err
	}
	s.storeCachedChallengeQuestions(ctx, QuestionTypePostReadingQuiz, inputHash, cacheKey, request, questions, aiModel, promptVersion)
	if err := s.challengeRepo.ReplaceByArticleAndType(ctx, articleID, QuestionTypePostReadingQuiz, questions); err != nil {
		return nil, err
	}
	return s.challengeRepo.ListByArticleAndType(ctx, articleID, QuestionTypePostReadingQuiz)
}

func (s *ChallengeService) ListPostQuizResults(ctx context.Context, userID, articleID int64) ([]model.ReadingAnswerDetail, error) {
	if _, err := s.articleRepo.GetAccessible(ctx, userID, articleID); err != nil {
		return nil, err
	}
	return s.challengeRepo.ListAttemptsByArticleAndType(ctx, userID, articleID, QuestionTypePostReadingQuiz)
}

func (s *ChallengeService) SubmitAnswer(ctx context.Context, userID, questionID int64, selectedOption string) (map[string]any, error) {
	selectedOption = strings.ToUpper(strings.TrimSpace(selectedOption))
	if !slices.Contains([]string{"A", "B", "C", "D"}, selectedOption) {
		return nil, fmt.Errorf("invalid selected option")
	}

	question, err := s.challengeRepo.GetAccessibleByID(ctx, userID, questionID)
	if err != nil {
		return nil, err
	}

	isCorrect := question.CorrectOption == selectedOption
	attempt := &model.ChallengeQuestionAttempt{
		QuestionID:     questionID,
		UserID:         userID,
		SelectedOption: selectedOption,
		IsCorrect:      isCorrect,
	}
	if _, err := s.challengeRepo.CreateAttempt(ctx, attempt); err != nil {
		return nil, err
	}

	return map[string]any{
		"is_correct":          isCorrect,
		"correct_option":      question.CorrectOption,
		"correct_answer_text": question.CorrectAnswerText,
		"explanation":         question.Explanation,
	}, nil
}

type aiQuestionResponse struct {
	Items []aiQuestionItem `json:"items"`
}

type aiQuestionItem struct {
	SentenceID        int64  `json:"sentence_id"`
	SentenceText      string `json:"sentence_text"`
	MaskedSentence    string `json:"masked_sentence"`
	CorrectAnswerText string `json:"correct_answer_text"`
	OptionA           string `json:"option_a"`
	OptionB           string `json:"option_b"`
	OptionC           string `json:"option_c"`
	OptionD           string `json:"option_d"`
	CorrectOption     string `json:"correct_option"`
	Explanation       string `json:"explanation"`
}

func (s *ChallengeService) generateQuestionsWithAI(
	ctx context.Context,
	articleID int64,
	request challengeQuestionCacheRequest,
	questionType string,
	aiModel string,
	promptVersion string,
) ([]model.ChallengeQuestion, error) {
	var prompt AIPrompt
	if questionType == QuestionTypePostReadingQuiz {
		prompt = promptPostQuizQuestions(request)
	} else {
		prompt = promptChallengeQuestions(request)
	}
	raw, err := s.aiService.CompleteJSON(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("generate reading questions with ai: %w", err)
	}
	var parsed aiQuestionResponse
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, fmt.Errorf("parse ai reading questions: %w", err)
	}
	if len(parsed.Items) == 0 {
		return nil, fmt.Errorf("ai did not return any reading questions")
	}

	sentenceByID := make(map[int64]challengeCacheSentence, len(request.Sentences))
	for _, sentence := range request.Sentences {
		sentenceByID[sentence.ID] = sentence
	}
	questions := make([]model.ChallengeQuestion, 0, min(len(parsed.Items), 8))
	for _, item := range parsed.Items {
		if len(questions) >= 8 {
			break
		}
		item.CorrectOption = strings.ToUpper(strings.TrimSpace(item.CorrectOption))
		options := []string{
			strings.TrimSpace(item.OptionA),
			strings.TrimSpace(item.OptionB),
			strings.TrimSpace(item.OptionC),
			strings.TrimSpace(item.OptionD),
		}
		if !slices.Contains([]string{"A", "B", "C", "D"}, item.CorrectOption) ||
			strings.TrimSpace(item.MaskedSentence) == "" ||
			strings.TrimSpace(item.CorrectAnswerText) == "" ||
			hasBlankOption(options) {
			continue
		}
		sentence, ok := sentenceByID[item.SentenceID]
		if !ok && len(request.Sentences) > 0 {
			sentence = request.Sentences[0]
		}
		sentenceText := strings.TrimSpace(item.SentenceText)
		if sentenceText == "" {
			sentenceText = sentence.SentenceText
		}
		entry, _, err := s.dictionarySvc.LookupOrGenerate(ctx, item.CorrectAnswerText)
		if err != nil {
			return nil, fmt.Errorf("ensure correct answer dictionary entry: %w", err)
		}
		questions = append(questions, model.ChallengeQuestion{
			ArticleID:         articleID,
			SentenceID:        sentence.ID,
			QuestionType:      questionType,
			QuestionOrder:     len(questions) + 1,
			SentenceText:      sentenceText,
			MaskedSentence:    strings.TrimSpace(item.MaskedSentence),
			CorrectEntryID:    entry.ID,
			CorrectAnswerText: strings.TrimSpace(item.CorrectAnswerText),
			OptionA:           options[0],
			OptionB:           options[1],
			OptionC:           options[2],
			OptionD:           options[3],
			CorrectOption:     item.CorrectOption,
			Explanation:       strings.TrimSpace(item.Explanation),
			JLPTLevel:         string(request.JLPTLevel),
			AIModel:           &aiModel,
			PromptVersion:     &promptVersion,
		})
	}
	if len(questions) == 0 {
		return nil, fmt.Errorf("ai returned no valid reading questions")
	}
	return questions, nil
}

func hasBlankOption(options []string) bool {
	seen := map[string]bool{}
	for _, option := range options {
		if strings.TrimSpace(option) == "" {
			return true
		}
		if seen[option] {
			return true
		}
		seen[option] = true
	}
	return false
}

func containsPlaceholderQuestions(questions []model.ChallengeQuestion) bool {
	for _, question := range questions {
		if question.AIModel != nil && strings.HasPrefix(*question.AIModel, "placeholder-") {
			return true
		}
		if strings.Contains(question.OptionA, "候补") || strings.Contains(question.OptionB, "候补") ||
			strings.Contains(question.OptionC, "候补") || strings.Contains(question.OptionD, "候补") ||
			strings.Contains(question.OptionA, "干扰项") || strings.Contains(question.OptionB, "干扰项") ||
			strings.Contains(question.OptionC, "干扰项") || strings.Contains(question.OptionD, "干扰项") {
			return true
		}
	}
	return false
}

func (s *ChallengeService) pickKeyword(ctx context.Context, sentence string, entries []model.DictionaryEntry) (string, *model.DictionaryEntry, error) {
	for _, entry := range entries {
		candidate := strings.TrimSpace(entry.Surface)
		if len([]rune(candidate)) >= 2 && strings.Contains(sentence, candidate) {
			entryCopy := entry
			return candidate, &entryCopy, nil
		}
	}

	token := extractJapaneseToken(sentence)
	if token == "" {
		return "", nil, nil
	}
	entry, _, err := s.dictionarySvc.LookupOrGenerate(ctx, token)
	return token, entry, err
}

func (s *ChallengeService) buildOptions(ctx context.Context, correct *model.DictionaryEntry) ([]string, string, string, error) {
	distractors, err := s.dictionaryRepo.ListDistractors(ctx, correct.ID, correct.PartOfSpeech, 10)
	if err != nil {
		return nil, "", "", err
	}

	optionSet := []string{correct.Surface}
	for _, entry := range distractors {
		if entry.Surface != "" && !slices.Contains(optionSet, entry.Surface) {
			optionSet = append(optionSet, entry.Surface)
		}
		if len(optionSet) == 4 {
			break
		}
	}
	for len(optionSet) < 4 {
		optionSet = append(optionSet, fmt.Sprintf("%s候补%d", correct.Surface, len(optionSet)))
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(correct.ID)))
	rng.Shuffle(len(optionSet), func(i, j int) {
		optionSet[i], optionSet[j] = optionSet[j], optionSet[i]
	})

	correctOption := "A"
	for idx, option := range optionSet {
		if option == correct.Surface {
			correctOption = string(rune('A' + idx))
			break
		}
	}
	explanation := fmt.Sprintf("正确答案是 %s，中文释义为：%s。", correct.Surface, correct.PrimaryMeaningZH)
	return optionSet, correctOption, explanation, nil
}

func (s *ChallengeService) buildMeaningOptions(ctx context.Context, correct *model.DictionaryEntry) ([]string, string, string, error) {
	distractors, err := s.dictionaryRepo.ListDistractors(ctx, correct.ID, correct.PartOfSpeech, 10)
	if err != nil {
		return nil, "", "", err
	}

	correctMeaning := strings.TrimSpace(correct.PrimaryMeaningZH)
	if correctMeaning == "" {
		correctMeaning = strings.TrimSpace(correct.MeaningZH)
	}
	optionSet := []string{correctMeaning}
	for _, entry := range distractors {
		meaning := strings.TrimSpace(entry.PrimaryMeaningZH)
		if meaning == "" {
			meaning = strings.TrimSpace(entry.MeaningZH)
		}
		if meaning != "" && !slices.Contains(optionSet, meaning) {
			optionSet = append(optionSet, meaning)
		}
		if len(optionSet) == 4 {
			break
		}
	}
	fallbacks := []string{"取消", "购买", "出发", "查看", "听见", "书写"}
	for _, fallback := range fallbacks {
		if len(optionSet) == 4 {
			break
		}
		if !slices.Contains(optionSet, fallback) {
			optionSet = append(optionSet, fallback)
		}
	}
	for len(optionSet) < 4 {
		optionSet = append(optionSet, fmt.Sprintf("干扰项%d", len(optionSet)))
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(correct.ID)*13))
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
	explanation := fmt.Sprintf("「%s」的主要中文意思是：%s。原句：%s", correct.Surface, correctMeaning, valueOrFallback(correct.ExampleSentence, "见文章原句"))
	return optionSet, correctOption, explanation, nil
}

type challengeQuestionCacheRequest struct {
	ArticleID    int64                    `json:"article_id"`
	JLPTLevel    model.JLPTLevel          `json:"jlpt_level"`
	QuestionType string                   `json:"question_type"`
	Sentences    []challengeCacheSentence `json:"sentences"`
}

type challengeCacheSentence struct {
	ID            int64  `json:"id"`
	SentenceOrder int    `json:"sentence_order"`
	SentenceText  string `json:"sentence_text"`
}

func challengeCacheRequest(article *model.Article, sentences []model.ArticleSentence, questionType string) challengeQuestionCacheRequest {
	request := challengeQuestionCacheRequest{
		ArticleID:    article.ID,
		JLPTLevel:    article.JLPTLevel,
		QuestionType: questionType,
		Sentences:    make([]challengeCacheSentence, 0, len(sentences)),
	}
	for _, sentence := range sentences {
		request.Sentences = append(request.Sentences, challengeCacheSentence{
			ID:            sentence.ID,
			SentenceOrder: sentence.SentenceOrder,
			SentenceText:  sentence.SentenceText,
		})
	}
	return request
}

func (s *ChallengeService) challengeCacheKey(request challengeQuestionCacheRequest, aiModel, promptVersion string) (string, string) {
	if s.aiService == nil {
		return "", ""
	}
	raw, _ := json.Marshal(request)
	inputHash := s.aiService.HashInput(string(raw))
	return s.aiService.CacheKey(request.QuestionType, inputHash, aiModel, promptVersion), inputHash
}

func (s *ChallengeService) getCachedChallengeQuestions(ctx context.Context, cacheKey string) ([]model.ChallengeQuestion, bool) {
	if s.aiService == nil || cacheKey == "" {
		return nil, false
	}
	cached, ok, err := s.aiService.GetCached(ctx, cacheKey)
	if err != nil || !ok {
		return nil, false
	}
	var questions []model.ChallengeQuestion
	if err := json.Unmarshal([]byte(cached.ResponseJSON), &questions); err != nil || len(questions) == 0 {
		return nil, false
	}
	for idx := range questions {
		questions[idx].ID = 0
		questions[idx].CreatedAt = time.Time{}
	}
	return questions, true
}

func (s *ChallengeService) storeCachedChallengeQuestions(ctx context.Context, taskType, inputHash, cacheKey string, request challengeQuestionCacheRequest, questions []model.ChallengeQuestion, aiModel, promptVersion string) {
	if s.aiService == nil || cacheKey == "" {
		return
	}
	cacheQuestions := make([]model.ChallengeQuestion, len(questions))
	copy(cacheQuestions, questions)
	for idx := range cacheQuestions {
		cacheQuestions[idx].ID = 0
		cacheQuestions[idx].CreatedAt = time.Time{}
	}
	if _, err := s.aiService.StoreCached(ctx, taskType, inputHash, cacheKey, request, cacheQuestions, aiModel, promptVersion); err != nil {
		s.aiService.LogFailure(ctx, taskType, request, err, aiModel, promptVersion)
	}
}

func valueOrFallback(value *string, fallback string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return fallback
	}
	return strings.TrimSpace(*value)
}

func extractJapaneseToken(sentence string) string {
	var current []rune
	var best string
	flush := func() {
		if len(current) >= 2 && len(current) <= 8 {
			candidate := string(current)
			if len([]rune(candidate)) > len([]rune(best)) {
				best = candidate
			}
		}
		current = current[:0]
	}

	for _, r := range sentence {
		if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
			current = append(current, r)
			continue
		}
		if len(current) > 0 {
			flush()
		}
	}
	if len(current) > 0 {
		flush()
	}
	return strings.TrimSpace(best)
}
