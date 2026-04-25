package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ai-japanese-learning/internal/model"
)

type TranslationService struct {
	aiService *AIService
}

func NewTranslationService(aiService *AIService) *TranslationService {
	return &TranslationService{aiService: aiService}
}

type translationResult struct {
	JapaneseContent string `json:"japanese_content"`
	SourceType      string `json:"source_type"`
	IsAIGenerated   bool   `json:"is_ai_generated"`
	Note            string `json:"note"`
}

func (s *TranslationService) TranslateToJapanese(ctx context.Context, language, content string, level model.JLPTLevel) (string, string, bool, string) {
	cleaned := strings.TrimSpace(content)
	if language == "ja" {
		return cleaned, "user_uploaded", false, "Japanese article detected. Stored without translation."
	}

	const (
		taskType      = "article_translation"
		modelName     = "placeholder-translation-service"
		promptVersion = "v0.8-review"
	)
	request := map[string]string{
		"language":   language,
		"content":    cleaned,
		"jlpt_level": string(level),
	}
	if s.aiService != nil {
		inputHash := s.aiService.HashInput(language + ":" + string(level) + ":" + cleaned)
		cacheKey := s.aiService.CacheKey(taskType, inputHash, modelName, promptVersion)
		if cached, ok, err := s.aiService.GetCached(ctx, cacheKey); err == nil && ok {
			var result translationResult
			if err := json.Unmarshal([]byte(cached.ResponseJSON), &result); err == nil && strings.TrimSpace(result.JapaneseContent) != "" {
				return result.JapaneseContent, result.SourceType, result.IsAIGenerated, result.Note
			}
		}

		result := buildPlaceholderTranslation(language, cleaned, level)
		if _, err := s.aiService.StoreCached(ctx, taskType, inputHash, cacheKey, request, result, modelName, promptVersion); err != nil {
			s.aiService.LogFailure(ctx, taskType, request, err, modelName, promptVersion)
		}
		return result.JapaneseContent, result.SourceType, result.IsAIGenerated, result.Note
	}

	result := buildPlaceholderTranslation(language, cleaned, level)
	return result.JapaneseContent, result.SourceType, result.IsAIGenerated, result.Note
}

func buildPlaceholderTranslation(language, cleaned string, level model.JLPTLevel) translationResult {
	translated := fmt.Sprintf(
		"これは %s 向けに整えた日本語版の記事です。\n\n原文要約:\n%s",
		level,
		cleaned,
	)
	return translationResult{
		JapaneseContent: translated,
		SourceType:      "ai_translated",
		IsAIGenerated:   true,
		Note:            "Using placeholder translation service with AI cache in v0.8-review. Replace with real AI provider in later versions.",
	}
}
