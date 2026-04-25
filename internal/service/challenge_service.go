package service

import (
	"context"
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
) *ChallengeService {
	return &ChallengeService{
		articleRepo:    articleRepo,
		dictionaryRepo: dictionaryRepo,
		challengeRepo:  challengeRepo,
		dictionarySvc:  dictionarySvc,
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
		return existing, nil
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

	dictionaryEntries, err := s.dictionaryRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	var questions []model.ChallengeQuestion
	for _, sentence := range sentences {
		keyword, entry, err := s.pickKeyword(ctx, sentence.SentenceText, dictionaryEntries)
		if err != nil || keyword == "" || entry == nil {
			continue
		}
		masked := strings.Replace(sentence.SentenceText, keyword, "____", 1)
		if masked == sentence.SentenceText {
			continue
		}

		options, correctOption, explanation, err := s.buildOptions(ctx, entry)
		if err != nil {
			return nil, err
		}
		aiModel := "placeholder-challenge-generator"
		promptVersion := "v0.5"

		questions = append(questions, model.ChallengeQuestion{
			ArticleID:         articleID,
			SentenceID:        sentence.ID,
			QuestionType:      QuestionTypeChallengeReading,
			QuestionOrder:     len(questions) + 1,
			SentenceText:      sentence.SentenceText,
			MaskedSentence:    masked,
			CorrectEntryID:    entry.ID,
			CorrectAnswerText: keyword,
			OptionA:           options[0],
			OptionB:           options[1],
			OptionC:           options[2],
			OptionD:           options[3],
			CorrectOption:     correctOption,
			Explanation:       explanation,
			JLPTLevel:         string(article.JLPTLevel),
			AIModel:           &aiModel,
			PromptVersion:     &promptVersion,
		})

		if len(questions) >= 5 {
			break
		}
	}

	if len(questions) == 0 {
		return nil, fmt.Errorf("no suitable challenge questions could be generated")
	}
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
		return existing, nil
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

	dictionaryEntries, err := s.dictionaryRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	var questions []model.ChallengeQuestion
	for _, sentence := range sentences {
		keyword, entry, err := s.pickKeyword(ctx, sentence.SentenceText, dictionaryEntries)
		if err != nil || keyword == "" || entry == nil {
			continue
		}

		options, correctOption, explanation, err := s.buildMeaningOptions(ctx, entry)
		if err != nil {
			return nil, err
		}
		aiModel := "placeholder-post-quiz-generator"
		promptVersion := "v0.6"
		prompt := fmt.Sprintf("读完文章后，请选择文中「%s」最合适的中文意思。", keyword)

		questions = append(questions, model.ChallengeQuestion{
			ArticleID:         articleID,
			SentenceID:        sentence.ID,
			QuestionType:      QuestionTypePostReadingQuiz,
			QuestionOrder:     len(questions) + 1,
			SentenceText:      sentence.SentenceText,
			MaskedSentence:    prompt,
			CorrectEntryID:    entry.ID,
			CorrectAnswerText: entry.PrimaryMeaningZH,
			OptionA:           options[0],
			OptionB:           options[1],
			OptionC:           options[2],
			OptionD:           options[3],
			CorrectOption:     correctOption,
			Explanation:       explanation,
			JLPTLevel:         string(article.JLPTLevel),
			AIModel:           &aiModel,
			PromptVersion:     &promptVersion,
		})

		if len(questions) >= 5 {
			break
		}
	}

	if len(questions) == 0 {
		return nil, fmt.Errorf("no suitable post-reading quiz questions could be generated")
	}
	if err := s.challengeRepo.ReplaceByArticleAndType(ctx, articleID, QuestionTypePostReadingQuiz, questions); err != nil {
		return nil, err
	}
	return s.challengeRepo.ListByArticleAndType(ctx, articleID, QuestionTypePostReadingQuiz)
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
