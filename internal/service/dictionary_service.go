package service

import (
	"context"
	"fmt"
	"strings"

	"ai-japanese-learning/internal/model"
	"ai-japanese-learning/internal/repository"
)

type DictionaryService struct {
	dictionaryRepo *repository.DictionaryRepository
}

func NewDictionaryService(dictionaryRepo *repository.DictionaryRepository) *DictionaryService {
	return &DictionaryService{dictionaryRepo: dictionaryRepo}
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

	generated := buildGeneratedDictionaryEntry(text)
	created, err := s.dictionaryRepo.Create(ctx, generated)
	if err != nil {
		return nil, false, err
	}
	return created, true, nil
}

func (s *DictionaryService) GetByID(ctx context.Context, entryID int64) (*model.DictionaryEntry, error) {
	return s.dictionaryRepo.GetByID(ctx, entryID)
}

func buildGeneratedDictionaryEntry(text string) *model.DictionaryEntry {
	meaningJA := fmt.Sprintf("%s の意味を後続バージョンで AI により補完予定です。", text)
	meaningEN := fmt.Sprintf("Placeholder entry generated for %s.", text)
	exampleSentence := fmt.Sprintf("記事内で「%s」が使われています。", text)
	exampleTranslation := fmt.Sprintf("The text \"%s\" appears in the article.", text)
	aiModel := "placeholder-dictionary-generator"
	promptVersion := "v0.3"
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
