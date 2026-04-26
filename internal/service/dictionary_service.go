package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ai-japanese-learning/internal/model"
	"ai-japanese-learning/internal/repository"
)

type DictionaryService struct {
	dictionaryRepo *repository.DictionaryRepository
	aiService      *AIService
}

type dictionaryExampleAIResponse struct {
	ExampleSentence      string `json:"example_sentence"`
	ExampleTranslationZH string `json:"example_translation_zh"`
}

func NewDictionaryService(dictionaryRepo *repository.DictionaryRepository, aiService *AIService) *DictionaryService {
	return &DictionaryService{
		dictionaryRepo: dictionaryRepo,
		aiService:      aiService,
	}
}

func (s *DictionaryService) Lookup(ctx context.Context, text string) (*model.DictionaryEntry, bool, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, false, fmt.Errorf("text is required")
	}

	entry, err := s.dictionaryRepo.FindByText(ctx, text)
	if err == nil {
		return entry, true, nil
	}
	if err == repository.ErrDictionaryEntryNotFound {
		return nil, false, nil
	}
	return nil, false, err
}

func (s *DictionaryService) LookupOrGenerate(ctx context.Context, text string) (*model.DictionaryEntry, bool, error) {
	return s.LookupOrGenerateWithContext(ctx, text, "")
}

func (s *DictionaryService) LookupOrGenerateWithContext(ctx context.Context, text string, contextText string) (*model.DictionaryEntry, bool, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, false, fmt.Errorf("text is required")
	}

	entry, err := s.dictionaryRepo.FindByText(ctx, text)
	if err == nil {
		return entry, false, nil
	}
	if err != nil && err != repository.ErrDictionaryEntryNotFound {
		return nil, false, err
	}

	generated, err := s.generateDictionaryEntry(ctx, text, contextText)
	if err != nil {
		return nil, false, err
	}
	originalForm := text
	if strings.TrimSpace(generated.Lemma) != "" {
		generated.Surface = strings.TrimSpace(generated.Lemma)
	}
	created, err := s.dictionaryRepo.Create(ctx, generated)
	if err != nil {
		return nil, false, err
	}
	if originalForm != created.Surface && originalForm != created.Lemma {
		if err := s.dictionaryRepo.CreateForm(ctx, created.ID, originalForm); err != nil {
			return nil, false, err
		}
	}
	return created, true, nil
}

func (s *DictionaryService) GetByID(ctx context.Context, entryID int64) (*model.DictionaryEntry, error) {
	return s.dictionaryRepo.GetByID(ctx, entryID)
}

func (s *DictionaryService) ListExamples(ctx context.Context, entryID int64) ([]model.DictionaryExample, error) {
	if _, err := s.dictionaryRepo.GetByID(ctx, entryID); err != nil {
		return nil, err
	}
	return s.dictionaryRepo.ListExamples(ctx, entryID)
}

func (s *DictionaryService) GenerateExample(ctx context.Context, entryID int64) (*model.DictionaryExample, error) {
	entry, err := s.dictionaryRepo.GetByID(ctx, entryID)
	if err != nil {
		return nil, err
	}
	existing, err := s.dictionaryRepo.ListExamples(ctx, entryID)
	if err != nil {
		return nil, err
	}
	if len(existing) >= 3 {
		return nil, fmt.Errorf("最多只能生成 3 句例句，请先删除旧例句")
	}

	const (
		taskType      = "dictionary_example"
		fallbackModel = "placeholder-dictionary-example-generator"
		promptVersion = aiPromptVersionV12
	)
	modelName := fallbackModel
	source := "ai"
	var parsed dictionaryExampleAIResponse
	if s.aiService != nil && s.aiService.ProviderAvailableFor(ctx) {
		modelName = s.aiService.ModelNameFor(ctx, fallbackModel)
		prompt := promptDictionaryExample(*entry, existing)
		request := map[string]any{"entry_id": entryID, "existing_count": len(existing), "prompt": prompt}
		raw, err := s.aiService.CompleteJSON(ctx, prompt)
		if err != nil {
			s.aiService.LogFailure(ctx, taskType, request, err, modelName, promptVersion)
		} else if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			s.aiService.LogFailure(ctx, taskType, request, err, modelName, promptVersion)
		}
	}
	if strings.TrimSpace(parsed.ExampleSentence) == "" {
		index := len(existing) + 1
		parsed.ExampleSentence = fmt.Sprintf("%sを使った例文%dです。", entry.Surface, index)
		parsed.ExampleTranslationZH = fmt.Sprintf("这是使用「%s」的例句 %d。", entry.Surface, index)
		modelName = fallbackModel
	}
	translation := strings.TrimSpace(parsed.ExampleTranslationZH)
	aiModel := modelName
	prompt := promptVersion
	return s.dictionaryRepo.CreateExample(ctx, &model.DictionaryExample{
		DictionaryEntryID:    entryID,
		ExampleSentence:      strings.TrimSpace(parsed.ExampleSentence),
		ExampleTranslationZH: &translation,
		Source:               source,
		AIModel:              &aiModel,
		PromptVersion:        &prompt,
	})
}

func (s *DictionaryService) DeleteExample(ctx context.Context, exampleID int64) error {
	if exampleID <= 0 {
		return fmt.Errorf("invalid example id")
	}
	return s.dictionaryRepo.DeleteExample(ctx, exampleID)
}

func (s *DictionaryService) generateDictionaryEntry(ctx context.Context, text string, contextText string) (*model.DictionaryEntry, error) {
	if s.aiService == nil {
		entry := buildGeneratedDictionaryEntry(text)
		return entry, validateDictionaryEntry(entry)
	}

	const (
		taskType      = "dictionary_entry"
		fallbackModel = "placeholder-dictionary-generator"
		promptVersion = aiPromptVersionV12
	)
	modelName := s.aiService.ModelNameFor(ctx, fallbackModel)
	prompt := promptDictionaryEntry(text, contextText)
	request := map[string]any{
		"text":           text,
		"context":        contextText,
		"prompt_version": promptVersion,
		"prompt":         prompt,
	}
	inputHash := s.aiService.HashInput(text + "\n" + contextText)
	cacheKey := s.aiService.CacheKey(taskType, inputHash, modelName, promptVersion)

	if cached, ok, err := s.aiService.GetCached(ctx, cacheKey); err != nil {
		return nil, err
	} else if ok {
		var entry model.DictionaryEntry
		if err := json.Unmarshal([]byte(cached.ResponseJSON), &entry); err != nil {
			s.aiService.LogFailure(ctx, taskType, request, err, modelName, promptVersion)
			return nil, fmt.Errorf("parse cached dictionary entry: %w", err)
		}
		if err := validateDictionaryEntry(&entry); err != nil {
			s.aiService.LogFailure(ctx, taskType, request, err, modelName, promptVersion)
			return nil, err
		}
		return &entry, nil
	}

	entry, err := s.generateDictionaryEntryWithAI(ctx, text, prompt, taskType, request, modelName, promptVersion)
	if err != nil {
		s.aiService.LogFailure(ctx, taskType, request, err, modelName, promptVersion)
		entry = buildGeneratedDictionaryEntry(text)
	} else {
		aiModel := modelName
		prompt := promptVersion
		entry.AIModel = &aiModel
		entry.PromptVersion = &prompt
	}
	aiModel := modelName
	promptName := promptVersion
	entry.AIModel = &aiModel
	entry.PromptVersion = &promptName
	if err := validateDictionaryEntry(entry); err != nil {
		s.aiService.LogFailure(ctx, taskType, request, err, modelName, promptVersion)
		return nil, err
	}
	if _, err := s.aiService.StoreCached(ctx, taskType, inputHash, cacheKey, request, entry, modelName, promptVersion); err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *DictionaryService) generateDictionaryEntryWithAI(ctx context.Context, text string, prompt AIPrompt, taskType string, request any, modelName, promptVersion string) (*model.DictionaryEntry, error) {
	if !s.aiService.ProviderAvailableFor(ctx) {
		return nil, fmt.Errorf("ai provider unavailable")
	}
	raw, err := s.aiService.CompleteJSON(ctx, prompt)
	if err != nil {
		return nil, err
	}
	var entry model.DictionaryEntry
	if err := json.Unmarshal([]byte(raw), &entry); err != nil {
		return nil, fmt.Errorf("parse ai dictionary entry: %w", err)
	}
	if strings.TrimSpace(entry.Surface) == "" {
		entry.Surface = text
	}
	entry.Source = "ai"
	entry.Verified = false
	if strings.TrimSpace(entry.ConfidenceScore) == "" {
		entry.ConfidenceScore = "0.80"
	}
	if err := validateDictionaryEntry(&entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func validateDictionaryEntry(entry *model.DictionaryEntry) error {
	if entry == nil {
		return fmt.Errorf("dictionary entry is nil")
	}
	if strings.TrimSpace(entry.Surface) == "" ||
		strings.TrimSpace(entry.Lemma) == "" ||
		strings.TrimSpace(entry.Reading) == "" ||
		strings.TrimSpace(entry.PartOfSpeech) == "" ||
		strings.TrimSpace(entry.MeaningZH) == "" ||
		strings.TrimSpace(entry.PrimaryMeaningZH) == "" ||
		strings.TrimSpace(entry.Source) == "" ||
		strings.TrimSpace(entry.ConfidenceScore) == "" {
		return fmt.Errorf("dictionary entry missing required fields")
	}
	switch entry.JLPTLevel {
	case "N5", "N4", "N3", "N2", "N1", "unknown":
	default:
		return fmt.Errorf("invalid dictionary jlpt level")
	}
	switch entry.Source {
	case "builtin", "ai", "admin":
	default:
		return fmt.Errorf("invalid dictionary source")
	}
	return nil
}

func buildGeneratedDictionaryEntry(text string) *model.DictionaryEntry {
	meaningJA := fmt.Sprintf("%s の意味を後続バージョンで AI により補完予定です。", text)
	meaningEN := fmt.Sprintf("Placeholder entry generated for %s.", text)
	exampleSentence := fmt.Sprintf("記事内で「%s」が使われています。", text)
	exampleTranslation := fmt.Sprintf("The text \"%s\" appears in the article.", text)
	aiModel := "placeholder-dictionary-generator"
	promptVersion := "v0.8"
	romaji := ""

	return &model.DictionaryEntry{
		Surface:              text,
		Lemma:                text,
		Reading:              text,
		Romaji:               &romaji,
		PartOfSpeech:         "unknown",
		MeaningZH:            fmt.Sprintf("%s：占位词条，后续版本将由 AI 生成更准确释义。", text),
		MeaningJA:            &meaningJA,
		MeaningEN:            &meaningEN,
		PrimaryMeaningZH:     fmt.Sprintf("%s（占位释义）", text),
		JLPTLevel:            "unknown",
		ExampleSentence:      &exampleSentence,
		ExampleTranslationZH: &exampleTranslation,
		IsCommon:             false,
		Source:               "ai",
		Verified:             false,
		ConfidenceScore:      "0.60",
		AIModel:              &aiModel,
		PromptVersion:        &promptVersion,
	}
}
