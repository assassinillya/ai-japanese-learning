package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"ai-japanese-learning/internal/model"
	"ai-japanese-learning/internal/repository"
)

type AIService struct {
	aiRepo   *repository.AIRepository
	provider AIProvider
}

func NewAIService(aiRepo *repository.AIRepository, provider AIProvider) *AIService {
	return &AIService{
		aiRepo:   aiRepo,
		provider: provider,
	}
}

func (s *AIService) CacheKey(taskType, inputHash, modelName, promptVersion string) string {
	return fmt.Sprintf("%s:%s:%s:%s", taskType, modelName, promptVersion, inputHash)
}

func (s *AIService) HashInput(input string) string {
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}

func (s *AIService) ProviderAvailable() bool {
	return s != nil && s.provider != nil
}

func (s *AIService) ModelName(fallback string) string {
	if s != nil && s.provider != nil && s.provider.ModelName() != "" {
		return s.provider.ModelName()
	}
	return fallback
}

func (s *AIService) CompleteJSON(ctx context.Context, prompt AIPrompt) (string, error) {
	if s == nil || s.provider == nil {
		return "", fmt.Errorf("ai provider is not configured")
	}
	return s.provider.CompleteJSON(ctx, prompt)
}

func (s *AIService) GetCached(ctx context.Context, cacheKey string) (*model.AICacheEntry, bool, error) {
	entry, err := s.aiRepo.GetCache(ctx, cacheKey)
	if err == nil {
		return entry, true, nil
	}
	if err == repository.ErrAICacheNotFound {
		return nil, false, nil
	}
	return nil, false, err
}

func (s *AIService) StoreCached(ctx context.Context, taskType, inputHash, cacheKey string, request any, response any, modelName, promptVersion string) (*model.AICacheEntry, error) {
	requestJSON, err := marshalJSONObject(request)
	if err != nil {
		return nil, err
	}
	responseJSON, err := marshalJSONObject(response)
	if err != nil {
		return nil, err
	}
	entry := &model.AICacheEntry{
		CacheKey:      cacheKey,
		TaskType:      taskType,
		InputHash:     inputHash,
		RequestJSON:   requestJSON,
		ResponseJSON:  responseJSON,
		ModelName:     modelName,
		PromptVersion: promptVersion,
	}
	if _, err := s.aiRepo.UpsertCache(ctx, entry); err != nil {
		return nil, err
	}
	responseJSONString := responseJSON
	_, _ = s.aiRepo.CreateLog(ctx, &model.AILog{
		TaskType:      taskType,
		RequestJSON:   requestJSON,
		ResponseJSON:  &responseJSONString,
		Status:        "success",
		ModelName:     modelName,
		PromptVersion: promptVersion,
	})
	return entry, nil
}

func (s *AIService) LogFailure(ctx context.Context, taskType string, request any, err error, modelName, promptVersion string) {
	requestJSON, marshalErr := marshalJSONObject(request)
	if marshalErr != nil {
		requestJSON = `{"error":"failed to marshal request"}`
	}
	message := err.Error()
	_, _ = s.aiRepo.CreateLog(ctx, &model.AILog{
		TaskType:      taskType,
		RequestJSON:   requestJSON,
		Status:        "failed",
		ErrorMessage:  &message,
		ModelName:     modelName,
		PromptVersion: promptVersion,
	})
}

func marshalJSONObject(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("marshal ai json: %w", err)
	}
	if !json.Valid(raw) {
		return "", fmt.Errorf("invalid ai json")
	}
	return string(raw), nil
}
