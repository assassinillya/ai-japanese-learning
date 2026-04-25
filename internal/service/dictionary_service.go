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

func NewDictionaryService(dictionaryRepo *repository.DictionaryRepository, aiService *AIService) *DictionaryService {
	return &DictionaryService{
		dictionaryRepo: dictionaryRepo,
		aiService:      aiService,
	}
}

func (s *DictionaryService) LookupOrGenerate(ctx context.Context, text string) (*model.DictionaryEntry, bool, error) {
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

	generated, err := s.generateDictionaryEntry(ctx, text)
	if err != nil {
		return nil, false, err
	}
	created, err := s.dictionaryRepo.Create(ctx, generated)
	if err != nil {
		return nil, false, err
	}
	return created, true, nil
}

func (s *DictionaryService) GetByID(ctx context.Context, entryID int64) (*model.DictionaryEntry, error) {
	return s.dictionaryRepo.GetByID(ctx, entryID)
}

func (s *DictionaryService) generateDictionaryEntry(ctx context.Context, text string) (*model.DictionaryEntry, error) {
	if s.aiService == nil {
		entry := buildGeneratedDictionaryEntry(text)
		return entry, validateDictionaryEntry(entry)
	}

	const (
		taskType      = "dictionary_entry"
		modelName     = "placeholder-dictionary-generator"
		promptVersion = "v0.8"
	)
	request := map[string]string{
		"text":           text,
		"prompt_version": promptVersion,
	}
	inputHash := s.aiService.HashInput(text)
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

	entry := buildGeneratedDictionaryEntry(text)
	aiModel := modelName
	prompt := promptVersion
	entry.AIModel = &aiModel
	entry.PromptVersion = &prompt
	if err := validateDictionaryEntry(entry); err != nil {
		s.aiService.LogFailure(ctx, taskType, request, err, modelName, promptVersion)
		return nil, err
	}
	if _, err := s.aiService.StoreCached(ctx, taskType, inputHash, cacheKey, request, entry, modelName, promptVersion); err != nil {
		return nil, err
	}
	return entry, nil
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
