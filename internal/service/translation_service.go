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
	ChineseSummary  string `json:"chinese_summary"`
	SourceType      string `json:"source_type"`
	IsAIGenerated   bool   `json:"is_ai_generated"`
	Note            string `json:"note"`
}

type articleSummaryResult struct {
	ChineseSummary string `json:"chinese_summary"`
}

func (s *TranslationService) TranslateToJapanese(ctx context.Context, language, content string, level model.JLPTLevel) (string, string, bool, string) {
	cleaned := strings.TrimSpace(content)
	if language == "ja" {
		return cleaned, "user_uploaded", false, "Japanese article detected. Stored without translation."
	}

	const (
		taskType      = "article_translation"
		fallbackModel = "placeholder-translation-service"
		promptVersion = aiPromptVersionV12
	)
	modelName := fallbackModel
	if s.aiService != nil {
		modelName = s.aiService.ModelNameFor(ctx, fallbackModel)
	}
	prompt := promptArticleTranslation(language, cleaned, level)
	request := map[string]any{
		"language":   language,
		"content":    cleaned,
		"jlpt_level": string(level),
		"prompt":     prompt,
	}
	if s.aiService != nil {
		inputHash := s.aiService.HashInput(language + ":" + string(level) + ":" + cleaned)
		cacheKey := s.aiService.CacheKey(taskType, inputHash, modelName, promptVersion)
		if cached, ok, err := s.aiService.GetCached(ctx, cacheKey); err == nil && ok {
			var result translationResult
			if err := json.Unmarshal([]byte(cached.ResponseJSON), &result); err == nil && strings.TrimSpace(result.JapaneseContent) != "" {
				return result.JapaneseContent, result.SourceType, result.IsAIGenerated, firstNonBlank(result.ChineseSummary, result.Note)
			}
		}

		result, err := s.translateWithAI(ctx, prompt)
		if err != nil {
			s.aiService.LogFailure(ctx, taskType, request, err, modelName, promptVersion)
			result = buildPlaceholderTranslation(language, cleaned, level)
		}
		if _, err := s.aiService.StoreCached(ctx, taskType, inputHash, cacheKey, request, result, modelName, promptVersion); err != nil {
			s.aiService.LogFailure(ctx, taskType, request, err, modelName, promptVersion)
		}
		return result.JapaneseContent, result.SourceType, result.IsAIGenerated, firstNonBlank(result.ChineseSummary, result.Note)
	}

	result := buildPlaceholderTranslation(language, cleaned, level)
	return result.JapaneseContent, result.SourceType, result.IsAIGenerated, firstNonBlank(result.ChineseSummary, result.Note)
}

func (s *TranslationService) SummarizeChinese(ctx context.Context, title, content string) string {
	title = strings.TrimSpace(title)
	cleaned := strings.TrimSpace(content)
	if cleaned == "" {
		return ""
	}
	if s.aiService == nil || !s.aiService.ProviderAvailableFor(ctx) {
		return buildLocalChineseSummary(cleaned)
	}

	const taskType = "article_chinese_summary"
	promptVersion := aiPromptVersionV12
	modelName := s.aiService.ModelNameFor(ctx, "ai-article-summary")
	prompt := promptArticleSummaryZH(title, cleaned)
	request := map[string]any{
		"title":          title,
		"content":        cleaned,
		"prompt":         prompt,
		"prompt_version": promptVersion,
	}
	inputHash := s.aiService.HashInput(title + "\n" + cleaned)
	cacheKey := s.aiService.CacheKey(taskType, inputHash, modelName, promptVersion)
	if cached, ok, err := s.aiService.GetCached(ctx, cacheKey); err == nil && ok {
		var result articleSummaryResult
		if err := json.Unmarshal([]byte(cached.ResponseJSON), &result); err == nil && strings.TrimSpace(result.ChineseSummary) != "" {
			return strings.TrimSpace(result.ChineseSummary)
		}
	}

	raw, err := s.aiService.CompleteJSON(ctx, prompt)
	if err != nil {
		s.aiService.LogFailure(ctx, taskType, request, err, modelName, promptVersion)
		return buildLocalChineseSummary(cleaned)
	}
	var result articleSummaryResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil || strings.TrimSpace(result.ChineseSummary) == "" {
		if err != nil {
			s.aiService.LogFailure(ctx, taskType, request, err, modelName, promptVersion)
		}
		return buildLocalChineseSummary(cleaned)
	}
	_, _ = s.aiService.StoreCached(ctx, taskType, inputHash, cacheKey, request, result, modelName, promptVersion)
	return strings.TrimSpace(result.ChineseSummary)
}

func (s *TranslationService) translateWithAI(ctx context.Context, prompt AIPrompt) (translationResult, error) {
	if !s.aiService.ProviderAvailableFor(ctx) {
		return translationResult{}, fmt.Errorf("ai provider unavailable")
	}
	raw, err := s.aiService.CompleteJSON(ctx, prompt)
	if err != nil {
		return translationResult{}, err
	}
	var result translationResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return translationResult{}, fmt.Errorf("parse ai translation: %w", err)
	}
	if strings.TrimSpace(result.JapaneseContent) == "" {
		return translationResult{}, fmt.Errorf("ai translation missing japanese_content")
	}
	if strings.TrimSpace(result.SourceType) == "" {
		result.SourceType = "ai_translated"
	}
	result.IsAIGenerated = true
	if strings.TrimSpace(result.ChineseSummary) == "" {
		result.ChineseSummary = buildLocalChineseSummary(result.JapaneseContent)
	}
	if strings.TrimSpace(result.Note) == "" {
		result.Note = "Generated by configured AI provider."
	}
	return result, nil
}

func buildPlaceholderTranslation(language, cleaned string, level model.JLPTLevel) translationResult {
	translated := fmt.Sprintf(
		"これは %s 向けに整えた日本語版の記事です。\n\n原文要約:\n%s",
		level,
		cleaned,
	)
	return translationResult{
		JapaneseContent: translated,
		ChineseSummary:  buildLocalChineseSummary(cleaned),
		SourceType:      "ai_translated",
		IsAIGenerated:   true,
		Note:            "Using placeholder translation service with AI cache in v0.8-review. Replace with real AI provider in later versions.",
	}
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func buildLocalChineseSummary(content string) string {
	cleaned := strings.Join(strings.Fields(content), " ")
	if cleaned == "" {
		return ""
	}
	runes := []rune(cleaned)
	if len(runes) > 72 {
		cleaned = string(runes[:72]) + "..."
	}
	return "这篇文章主要围绕“" + cleaned + "”展开，适合作为日语阅读材料。"
}
