package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"
)

const (
	AIProviderPlaceholder     = "placeholder"
	AIProviderOpenAI          = "openai"
	AIProviderOpenAIResponses = "openai-responses"
	AIProviderGemini          = "gemini"
	AIProviderAnthropic       = "anthropic"
	AIProviderAzureOpenAI     = "azure-openai"
	AIProviderNewAPI          = "new-api"
)

type AIPrompt struct {
	System string `json:"system"`
	User   string `json:"user"`
}

type AIProviderConfig struct {
	Provider     string `json:"provider"`
	ProviderName string `json:"provider_name"`
	BaseURL      string `json:"base_url"`
	APIKey       string `json:"api_key,omitempty"`
	Model        string `json:"model"`
	APIVersion   string `json:"api_version,omitempty"`
}

type AIProviderDefinition struct {
	Type        string `json:"type"`
	Label       string `json:"label"`
	DefaultURL  string `json:"default_url"`
	DefaultPath string `json:"default_path"`
	ModelsPath  string `json:"models_path"`
	NeedsModel  bool   `json:"needs_model"`
}

type AIProviderStatus struct {
	Provider       string   `json:"provider"`
	ProviderName   string   `json:"provider_name"`
	BaseURL        string   `json:"base_url"`
	Model          string   `json:"model"`
	APIVersion     string   `json:"api_version,omitempty"`
	APIKeySaved    bool     `json:"api_key_saved"`
	Configured     bool     `json:"configured"`
	Endpoint       string   `json:"endpoint"`
	ModelsEndpoint string   `json:"models_endpoint"`
	Supported      []string `json:"supported"`
}

type AIProvider interface {
	ModelName() string
	CompleteJSON(ctx context.Context, prompt AIPrompt) (string, error)
	ListModels(ctx context.Context) ([]string, error)
	Check(ctx context.Context) error
	Status() AIProviderStatus
}

func AIProviderDefinitions() []AIProviderDefinition {
	return []AIProviderDefinition{
		{Type: AIProviderOpenAI, Label: "OpenAI", DefaultURL: "https://api.openai.com", DefaultPath: "/v1/chat/completions", ModelsPath: "/v1/models", NeedsModel: true},
		{Type: AIProviderOpenAIResponses, Label: "OpenAI Responses", DefaultURL: "https://api.openai.com", DefaultPath: "/v1/responses", ModelsPath: "/v1/models", NeedsModel: true},
		{Type: AIProviderGemini, Label: "Gemini", DefaultURL: "https://generativelanguage.googleapis.com", DefaultPath: "/v1beta/models/{model}:generateContent", ModelsPath: "/v1beta/models", NeedsModel: true},
		{Type: AIProviderAnthropic, Label: "Anthropic", DefaultURL: "https://api.anthropic.com", DefaultPath: "/v1/messages", ModelsPath: "/v1/models", NeedsModel: true},
		{Type: AIProviderAzureOpenAI, Label: "Azure OpenAI", DefaultURL: "https://{resource}.openai.azure.com", DefaultPath: "/openai/deployments/{model}/chat/completions", ModelsPath: "/openai/models", NeedsModel: true},
		{Type: AIProviderNewAPI, Label: "New API", DefaultURL: "https://your-new-api.example.com", DefaultPath: "/v1/chat/completions", ModelsPath: "/v1/models", NeedsModel: true},
	}
}

func NewAIProvider(provider, baseURL, apiKey, model string) AIProvider {
	return NewAIProviderFromConfig(AIProviderConfig{
		Provider: provider,
		BaseURL:  baseURL,
		APIKey:   apiKey,
		Model:    model,
	})
}

func NewAIProviderFromConfig(cfg AIProviderConfig) AIProvider {
	cfg = NormalizeAIProviderConfig(cfg)
	if strings.TrimSpace(cfg.APIKey) == "" || cfg.Provider == AIProviderPlaceholder {
		return nil
	}
	switch cfg.Provider {
	case AIProviderOpenAI, AIProviderNewAPI:
		return newOpenAIChatProvider(cfg)
	case AIProviderOpenAIResponses:
		return newOpenAIResponsesProvider(cfg)
	case AIProviderGemini:
		return newGeminiProvider(cfg)
	case AIProviderAnthropic:
		return newAnthropicProvider(cfg)
	case AIProviderAzureOpenAI:
		return newAzureOpenAIProvider(cfg)
	default:
		return nil
	}
}

func NormalizeAIProviderConfig(cfg AIProviderConfig) AIProviderConfig {
	cfg.Provider = normalizeProviderType(cfg.Provider)
	cfg.ProviderName = strings.TrimSpace(cfg.ProviderName)
	cfg.BaseURL = strings.TrimSpace(cfg.BaseURL)
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	cfg.Model = strings.TrimSpace(cfg.Model)
	cfg.APIVersion = strings.TrimSpace(cfg.APIVersion)
	if cfg.ProviderName == "" {
		cfg.ProviderName = providerLabel(cfg.Provider)
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultProviderURL(cfg.Provider)
	}
	if cfg.Model == "" {
		cfg.Model = defaultProviderModel(cfg.Provider)
	}
	if cfg.Provider == AIProviderAzureOpenAI && cfg.APIVersion == "" {
		cfg.APIVersion = "2024-10-21"
	}
	return cfg
}

func SanitizedAIProviderStatus(cfg AIProviderConfig, provider AIProvider) AIProviderStatus {
	cfg = NormalizeAIProviderConfig(cfg)
	status := AIProviderStatus{
		Provider:       cfg.Provider,
		ProviderName:   cfg.ProviderName,
		BaseURL:        cfg.BaseURL,
		Model:          cfg.Model,
		APIVersion:     cfg.APIVersion,
		APIKeySaved:    strings.TrimSpace(cfg.APIKey) != "",
		Configured:     provider != nil,
		Endpoint:       providerEndpoint(cfg),
		ModelsEndpoint: providerModelsEndpoint(cfg),
		Supported:      supportedProviderTypes(),
	}
	return status
}

type openAIChatProvider struct {
	cfg    AIProviderConfig
	client *http.Client
}

func newOpenAIChatProvider(cfg AIProviderConfig) *openAIChatProvider {
	return &openAIChatProvider{cfg: cfg, client: defaultAIHTTPClient()}
}

func (p *openAIChatProvider) ModelName() string { return p.cfg.Model }

func (p *openAIChatProvider) Status() AIProviderStatus {
	return SanitizedAIProviderStatus(p.cfg, p)
}

func (p *openAIChatProvider) CompleteJSON(ctx context.Context, prompt AIPrompt) (string, error) {
	payload := map[string]any{
		"model": p.cfg.Model,
		"messages": []map[string]string{
			{"role": "system", "content": prompt.System},
			{"role": "user", "content": prompt.User},
		},
		"temperature": 0.2,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}
	body, err := p.postJSON(ctx, providerEndpoint(p.cfg), payload)
	if err != nil {
		return "", err
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
	return requireJSONObject(parsed.Choices[0].Message.Content)
}

func (p *openAIChatProvider) ListModels(ctx context.Context) ([]string, error) {
	body, err := p.getJSON(ctx, providerModelsEndpoint(p.cfg))
	if err != nil {
		return nil, err
	}
	return parseModelIDs(body)
}

func (p *openAIChatProvider) Check(ctx context.Context) error {
	_, err := p.ListModels(ctx)
	return err
}

func (p *openAIChatProvider) postJSON(ctx context.Context, endpoint string, payload any) ([]byte, error) {
	return doJSON(ctx, p.client, http.MethodPost, endpoint, p.cfg, payload)
}

func (p *openAIChatProvider) getJSON(ctx context.Context, endpoint string) ([]byte, error) {
	return doJSON(ctx, p.client, http.MethodGet, endpoint, p.cfg, nil)
}

type openAIResponsesProvider struct {
	cfg    AIProviderConfig
	client *http.Client
}

func newOpenAIResponsesProvider(cfg AIProviderConfig) *openAIResponsesProvider {
	return &openAIResponsesProvider{cfg: cfg, client: defaultAIHTTPClient()}
}

func (p *openAIResponsesProvider) ModelName() string { return p.cfg.Model }

func (p *openAIResponsesProvider) Status() AIProviderStatus {
	return SanitizedAIProviderStatus(p.cfg, p)
}

func (p *openAIResponsesProvider) CompleteJSON(ctx context.Context, prompt AIPrompt) (string, error) {
	payload := map[string]any{
		"model": p.cfg.Model,
		"input": []map[string]any{
			{"role": "system", "content": prompt.System},
			{"role": "user", "content": prompt.User},
		},
		"temperature": 0.2,
		"text": map[string]any{
			"format": map[string]string{"type": "json_object"},
		},
	}
	body, err := doJSON(ctx, p.client, http.MethodPost, providerEndpoint(p.cfg), p.cfg, payload)
	if err != nil {
		return "", err
	}
	if content := extractString(body, "output_text"); content != "" {
		return requireJSONObject(content)
	}
	return "", fmt.Errorf("ai response has no output_text")
}

func (p *openAIResponsesProvider) ListModels(ctx context.Context) ([]string, error) {
	body, err := doJSON(ctx, p.client, http.MethodGet, providerModelsEndpoint(p.cfg), p.cfg, nil)
	if err != nil {
		return nil, err
	}
	return parseModelIDs(body)
}

func (p *openAIResponsesProvider) Check(ctx context.Context) error {
	_, err := p.ListModels(ctx)
	return err
}

type geminiProvider struct {
	cfg    AIProviderConfig
	client *http.Client
}

func newGeminiProvider(cfg AIProviderConfig) *geminiProvider {
	return &geminiProvider{cfg: cfg, client: defaultAIHTTPClient()}
}

func (p *geminiProvider) ModelName() string { return p.cfg.Model }

func (p *geminiProvider) Status() AIProviderStatus {
	return SanitizedAIProviderStatus(p.cfg, p)
}

func (p *geminiProvider) CompleteJSON(ctx context.Context, prompt AIPrompt) (string, error) {
	payload := map[string]any{
		"systemInstruction": map[string]any{
			"parts": []map[string]string{{"text": prompt.System}},
		},
		"contents": []map[string]any{
			{
				"role":  "user",
				"parts": []map[string]string{{"text": prompt.User}},
			},
		},
		"generationConfig": map[string]any{
			"temperature":      0.2,
			"responseMimeType": "application/json",
		},
	}
	body, err := doJSON(ctx, p.client, http.MethodPost, withGeminiKey(providerEndpoint(p.cfg), p.cfg.APIKey), p.cfg, payload)
	if err != nil {
		return "", err
	}
	var parsed struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("parse gemini response: %w", err)
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini response has no content")
	}
	return requireJSONObject(parsed.Candidates[0].Content.Parts[0].Text)
}

func (p *geminiProvider) ListModels(ctx context.Context) ([]string, error) {
	body, err := doJSON(ctx, p.client, http.MethodGet, withGeminiKey(providerModelsEndpoint(p.cfg), p.cfg.APIKey), p.cfg, nil)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("parse gemini models: %w", err)
	}
	models := make([]string, 0, len(parsed.Models))
	for _, model := range parsed.Models {
		name := strings.TrimPrefix(strings.TrimSpace(model.Name), "models/")
		if name != "" {
			models = append(models, name)
		}
	}
	sort.Strings(models)
	return models, nil
}

func (p *geminiProvider) Check(ctx context.Context) error {
	_, err := p.ListModels(ctx)
	return err
}

type anthropicProvider struct {
	cfg    AIProviderConfig
	client *http.Client
}

func newAnthropicProvider(cfg AIProviderConfig) *anthropicProvider {
	return &anthropicProvider{cfg: cfg, client: defaultAIHTTPClient()}
}

func (p *anthropicProvider) ModelName() string { return p.cfg.Model }

func (p *anthropicProvider) Status() AIProviderStatus {
	return SanitizedAIProviderStatus(p.cfg, p)
}

func (p *anthropicProvider) CompleteJSON(ctx context.Context, prompt AIPrompt) (string, error) {
	payload := map[string]any{
		"model":       p.cfg.Model,
		"system":      prompt.System,
		"max_tokens":  4096,
		"temperature": 0.2,
		"messages": []map[string]string{
			{"role": "user", "content": prompt.User},
		},
	}
	body, err := doJSON(ctx, p.client, http.MethodPost, providerEndpoint(p.cfg), p.cfg, payload)
	if err != nil {
		return "", err
	}
	var parsed struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("parse anthropic response: %w", err)
	}
	for _, content := range parsed.Content {
		if strings.TrimSpace(content.Text) != "" {
			return requireJSONObject(content.Text)
		}
	}
	return "", fmt.Errorf("anthropic response has no text content")
}

func (p *anthropicProvider) ListModels(ctx context.Context) ([]string, error) {
	body, err := doJSON(ctx, p.client, http.MethodGet, providerModelsEndpoint(p.cfg), p.cfg, nil)
	if err != nil {
		return nil, err
	}
	return parseModelIDs(body)
}

func (p *anthropicProvider) Check(ctx context.Context) error {
	_, err := p.ListModels(ctx)
	return err
}

type azureOpenAIProvider struct {
	cfg    AIProviderConfig
	client *http.Client
}

func newAzureOpenAIProvider(cfg AIProviderConfig) *azureOpenAIProvider {
	return &azureOpenAIProvider{cfg: cfg, client: defaultAIHTTPClient()}
}

func (p *azureOpenAIProvider) ModelName() string { return p.cfg.Model }

func (p *azureOpenAIProvider) Status() AIProviderStatus {
	return SanitizedAIProviderStatus(p.cfg, p)
}

func (p *azureOpenAIProvider) CompleteJSON(ctx context.Context, prompt AIPrompt) (string, error) {
	chat := newOpenAIChatProvider(p.cfg)
	chat.client = p.client
	return chat.CompleteJSON(ctx, prompt)
}

func (p *azureOpenAIProvider) ListModels(ctx context.Context) ([]string, error) {
	body, err := doJSON(ctx, p.client, http.MethodGet, providerModelsEndpoint(p.cfg), p.cfg, nil)
	if err != nil {
		return nil, err
	}
	return parseModelIDs(body)
}

func (p *azureOpenAIProvider) Check(ctx context.Context) error {
	if p.cfg.Model == "" {
		return fmt.Errorf("azure openai requires deployment name in model field")
	}
	_, err := p.ListModels(ctx)
	return err
}

func normalizeProviderType(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", AIProviderPlaceholder, "none":
		return AIProviderPlaceholder
	case "openai", "openai-chat", "openai-compatible", "compatible":
		return AIProviderOpenAI
	case "openai-response", "openai-responses", "responses", "response":
		return AIProviderOpenAIResponses
	case "gemini", "google", "google-gemini":
		return AIProviderGemini
	case "anthropic", "claude":
		return AIProviderAnthropic
	case "azure", "azure-openai", "azure-openal":
		return AIProviderAzureOpenAI
	case "newapi", "new-api":
		return AIProviderNewAPI
	default:
		return strings.ToLower(strings.TrimSpace(provider))
	}
}

func providerLabel(provider string) string {
	for _, definition := range AIProviderDefinitions() {
		if definition.Type == provider {
			return definition.Label
		}
	}
	return "未配置"
}

func defaultProviderURL(provider string) string {
	for _, definition := range AIProviderDefinitions() {
		if definition.Type == provider {
			return definition.DefaultURL
		}
	}
	return ""
}

func defaultProviderModel(provider string) string {
	switch provider {
	case AIProviderOpenAI, AIProviderOpenAIResponses, AIProviderNewAPI:
		return "gpt-4o-mini"
	case AIProviderGemini:
		return "gemini-1.5-flash"
	case AIProviderAnthropic:
		return "claude-3-5-haiku-latest"
	default:
		return ""
	}
}

func supportedProviderTypes() []string {
	definitions := AIProviderDefinitions()
	values := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		values = append(values, definition.Type)
	}
	return values
}

func providerEndpoint(cfg AIProviderConfig) string {
	cfg = NormalizeAIProviderConfig(cfg)
	switch cfg.Provider {
	case AIProviderOpenAIResponses:
		return appendEndpointSuffixAliases(cfg.BaseURL, "/v1/responses", "/v1/response")
	case AIProviderGemini:
		model := strings.TrimPrefix(cfg.Model, "models/")
		return appendEndpointSuffix(cfg.BaseURL, "/v1beta/models/"+url.PathEscape(model)+":generateContent")
	case AIProviderAnthropic:
		return appendEndpointSuffix(cfg.BaseURL, "/v1/messages")
	case AIProviderAzureOpenAI:
		endpoint := appendEndpointSuffix(cfg.BaseURL, "/openai/deployments/"+url.PathEscape(cfg.Model)+"/chat/completions")
		return withQuery(endpoint, "api-version", cfg.APIVersion)
	default:
		return appendEndpointSuffix(cfg.BaseURL, "/v1/chat/completions")
	}
}

func providerModelsEndpoint(cfg AIProviderConfig) string {
	cfg = NormalizeAIProviderConfig(cfg)
	switch cfg.Provider {
	case AIProviderGemini:
		return appendEndpointSuffix(cfg.BaseURL, "/v1beta/models")
	case AIProviderAnthropic:
		return appendEndpointSuffix(cfg.BaseURL, "/v1/models")
	case AIProviderAzureOpenAI:
		endpoint := appendEndpointSuffix(cfg.BaseURL, "/openai/models")
		return withQuery(endpoint, "api-version", cfg.APIVersion)
	default:
		return appendEndpointSuffix(cfg.BaseURL, "/v1/models")
	}
}

func appendEndpointSuffix(rawBase, suffix string) string {
	return appendEndpointSuffixAliases(rawBase, suffix)
}

func appendEndpointSuffixAliases(rawBase, suffix string, aliases ...string) string {
	rawBase = strings.TrimSpace(rawBase)
	suffix = "/" + strings.Trim(strings.TrimSpace(suffix), "/")
	if rawBase == "" {
		return suffix
	}
	parsed, err := url.Parse(rawBase)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return strings.TrimRight(rawBase, "/") + suffix
	}
	originalPath := strings.TrimRight(parsed.Path, "/")
	suffixPath := strings.TrimRight(suffix, "/")
	allSuffixes := append([]string{suffixPath}, aliases...)
	for _, candidate := range allSuffixes {
		candidate = "/" + strings.Trim(strings.TrimSpace(candidate), "/")
		if strings.EqualFold(originalPath, candidate) || strings.HasSuffix(strings.ToLower(originalPath), strings.ToLower(candidate)) {
			return parsed.String()
		}
	}
	if strings.EqualFold(originalPath, suffixPath) || strings.HasSuffix(strings.ToLower(originalPath), strings.ToLower(suffixPath)) {
		return parsed.String()
	}
	suffixParts := strings.Split(strings.Trim(suffixPath, "/"), "/")
	pathParts := strings.Split(strings.Trim(originalPath, "/"), "/")
	overlap := 0
	for i := 1; i <= len(suffixParts) && i <= len(pathParts); i++ {
		if equalStringSliceFold(pathParts[len(pathParts)-i:], suffixParts[:i]) {
			overlap = i
		}
	}
	addParts := suffixParts[overlap:]
	joined := originalPath
	if len(addParts) > 0 {
		joined = path.Join(append([]string{"/" + strings.Trim(originalPath, "/")}, addParts...)...)
	}
	if joined == "." || joined == "/" {
		joined = suffixPath
	}
	parsed.Path = joined
	return parsed.String()
}

func equalStringSliceFold(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !strings.EqualFold(a[i], b[i]) {
			return false
		}
	}
	return true
}

func withQuery(endpoint, key, value string) string {
	if strings.TrimSpace(value) == "" {
		return endpoint
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return endpoint
	}
	query := parsed.Query()
	if query.Get(key) == "" {
		query.Set(key, value)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func withGeminiKey(endpoint, apiKey string) string {
	return withQuery(endpoint, "key", apiKey)
}

func defaultAIHTTPClient() *http.Client {
	return &http.Client{Timeout: 45 * time.Second}
}

func doJSON(ctx context.Context, client *http.Client, method, endpoint string, cfg AIProviderConfig, payload any) ([]byte, error) {
	var reader io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal ai request: %w", err)
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return nil, fmt.Errorf("create ai request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	setAuthHeaders(req, cfg)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call ai provider: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("read ai response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ai provider status %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func setAuthHeaders(req *http.Request, cfg AIProviderConfig) {
	switch cfg.Provider {
	case AIProviderAnthropic:
		req.Header.Set("x-api-key", cfg.APIKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	case AIProviderAzureOpenAI:
		req.Header.Set("api-key", cfg.APIKey)
	case AIProviderGemini:
		if !strings.Contains(req.URL.RawQuery, "key=") {
			req.Header.Set("x-goog-api-key", cfg.APIKey)
		}
	default:
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}
}

func parseModelIDs(body []byte) ([]string, error) {
	var parsed struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("parse models: %w", err)
	}
	models := make([]string, 0, len(parsed.Data))
	for _, model := range parsed.Data {
		if strings.TrimSpace(model.ID) != "" {
			models = append(models, strings.TrimSpace(model.ID))
		}
	}
	sort.Strings(models)
	return models, nil
}

func requireJSONObject(content string) (string, error) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)
	if !json.Valid([]byte(content)) {
		return "", fmt.Errorf("ai response content is not valid json")
	}
	return content, nil
}

func extractString(body []byte, key string) string {
	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		return ""
	}
	return findStringValue(value, key)
}

func findStringValue(value any, key string) string {
	switch typed := value.(type) {
	case map[string]any:
		for k, v := range typed {
			if k == key {
				if str, ok := v.(string); ok {
					return strings.TrimSpace(str)
				}
			}
			if nested := findStringValue(v, key); nested != "" {
				return nested
			}
		}
	case []any:
		for _, item := range typed {
			if nested := findStringValue(item, key); nested != "" {
				return nested
			}
		}
	}
	return ""
}
