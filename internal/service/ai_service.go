package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"ai-japanese-learning/internal/model"
	"ai-japanese-learning/internal/repository"
)

type AIService struct {
	aiRepo   *repository.AIRepository
	mu       sync.RWMutex
	config   AIProviderConfig
	provider AIProvider
}

type aiContextKey struct{}

func ContextWithAIProviderConfig(ctx context.Context, cfg AIProviderConfig) context.Context {
	cfg = NormalizeAIProviderConfig(cfg)
	if NewAIProviderFromConfig(cfg) == nil {
		return ctx
	}
	return context.WithValue(ctx, aiContextKey{}, cfg)
}

func aiProviderConfigFromContext(ctx context.Context) (AIProviderConfig, bool) {
	cfg, ok := ctx.Value(aiContextKey{}).(AIProviderConfig)
	return cfg, ok
}

func NewAIService(aiRepo *repository.AIRepository, provider AIProvider) *AIService {
	return &AIService{
		aiRepo:   aiRepo,
		provider: provider,
	}
}

func NewConfiguredAIService(aiRepo *repository.AIRepository, cfg AIProviderConfig) *AIService {
	cfg = NormalizeAIProviderConfig(cfg)
	return &AIService{
		aiRepo:   aiRepo,
		config:   cfg,
		provider: NewAIProviderFromConfig(cfg),
	}
}

func (s *AIService) CurrentStatus() AIProviderStatus {
	if s == nil {
		return SanitizedAIProviderStatus(AIProviderConfig{}, nil)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return SanitizedAIProviderStatus(s.config, s.provider)
}

func (s *AIService) StatusForConfig(cfg AIProviderConfig) AIProviderStatus {
	return SanitizedAIProviderStatus(cfg, NewAIProviderFromConfig(cfg))
}

func (s *AIService) ConfigureProvider(cfg AIProviderConfig) (AIProviderStatus, error) {
	if s == nil {
		return AIProviderStatus{}, fmt.Errorf("ai service is not initialized")
	}
	cfg = NormalizeAIProviderConfig(cfg)
	provider := NewAIProviderFromConfig(cfg)
	s.mu.Lock()
	s.config = cfg
	s.provider = provider
	s.mu.Unlock()
	return SanitizedAIProviderStatus(cfg, provider), nil
}

func (s *AIService) ListProviderModels(ctx context.Context, cfg AIProviderConfig) ([]string, AIProviderStatus, error) {
	cfg = NormalizeAIProviderConfig(cfg)
	provider := NewAIProviderFromConfig(cfg)
	status := SanitizedAIProviderStatus(cfg, provider)
	if provider == nil {
		return nil, status, fmt.Errorf("api key is required for %s", status.ProviderName)
	}
	models, err := provider.ListModels(ctx)
	return models, status, err
}

func (s *AIService) CheckProvider(ctx context.Context, cfg AIProviderConfig) (AIProviderStatus, error) {
	cfg = NormalizeAIProviderConfig(cfg)
	provider := NewAIProviderFromConfig(cfg)
	status := SanitizedAIProviderStatus(cfg, provider)
	if provider == nil {
		return status, fmt.Errorf("api key is required for %s", status.ProviderName)
	}
	return status, provider.Check(ctx)
}

func (s *AIService) currentProvider() AIProvider {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.provider
}

func (s *AIService) providerForContext(ctx context.Context) AIProvider {
	if cfg, ok := aiProviderConfigFromContext(ctx); ok {
		return NewAIProviderFromConfig(cfg)
	}
	return s.currentProvider()
}

func (s *AIService) CacheKey(taskType, inputHash, modelName, promptVersion string) string {
	return fmt.Sprintf("%s:%s:%s:%s", taskType, modelName, promptVersion, inputHash)
}

func (s *AIService) HashInput(input string) string {
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}

func (s *AIService) ProviderAvailable() bool {
	return s != nil && s.currentProvider() != nil
}

func (s *AIService) ProviderAvailableFor(ctx context.Context) bool {
	return s != nil && s.providerForContext(ctx) != nil
}

func (s *AIService) ModelName(fallback string) string {
	provider := s.currentProvider()
	if provider != nil && provider.ModelName() != "" {
		return provider.ModelName()
	}
	return fallback
}

func (s *AIService) ModelNameFor(ctx context.Context, fallback string) string {
	provider := s.providerForContext(ctx)
	if provider != nil && provider.ModelName() != "" {
		return provider.ModelName()
	}
	return fallback
}

func (s *AIService) CompleteJSON(ctx context.Context, prompt AIPrompt) (string, error) {
	provider := s.providerForContext(ctx)
	if provider == nil {
		return "", fmt.Errorf("ai provider is not configured")
	}
	return provider.CompleteJSON(ctx, prompt)
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
