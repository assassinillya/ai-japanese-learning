package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type AIPrompt struct {
	System string `json:"system"`
	User   string `json:"user"`
}

type AIProvider interface {
	ModelName() string
	CompleteJSON(ctx context.Context, prompt AIPrompt) (string, error)
}

func NewAIProvider(provider, baseURL, apiKey, model string) AIProvider {
	if strings.TrimSpace(apiKey) == "" {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "openai", "openai-compatible", "compatible":
		return NewOpenAICompatibleProvider(baseURL, apiKey, model)
	default:
		return nil
	}
}

type OpenAICompatibleProvider struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

func NewOpenAICompatibleProvider(baseURL, apiKey, model string) *OpenAICompatibleProvider {
	return &OpenAICompatibleProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
		client: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (p *OpenAICompatibleProvider) ModelName() string {
	return p.model
}

func (p *OpenAICompatibleProvider) CompleteJSON(ctx context.Context, prompt AIPrompt) (string, error) {
	payload := map[string]any{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "system", "content": prompt.System},
			{"role": "user", "content": prompt.User},
		},
		"temperature": 0.2,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal ai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return "", fmt.Errorf("create ai request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call ai provider: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", fmt.Errorf("read ai response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("ai provider status %d: %s", resp.StatusCode, string(body))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("parse ai response envelope: %w", err)
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("ai response has no content")
	}
	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if !json.Valid([]byte(content)) {
		return "", fmt.Errorf("ai response content is not valid json")
	}
	return content, nil
}
