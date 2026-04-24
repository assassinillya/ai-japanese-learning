package service

import (
	"fmt"
	"strings"

	"ai-japanese-learning/internal/model"
)

type TranslationService struct{}

func NewTranslationService() *TranslationService {
	return &TranslationService{}
}

func (s *TranslationService) TranslateToJapanese(language, content string, level model.JLPTLevel) (string, string, bool, string) {
	cleaned := strings.TrimSpace(content)
	if language == "ja" {
		return cleaned, "user_uploaded", false, "Japanese article detected. Stored without translation."
	}

	translated := fmt.Sprintf(
		"これは %s 向けに整えた日本語版の記事です。\n\n原文要約:\n%s",
		level,
		cleaned,
	)
	return translated, "ai_translated", true, "Using placeholder translation service in v0.2. Replace with real AI provider in later versions."
}
