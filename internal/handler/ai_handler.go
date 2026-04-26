package handler

import (
	"encoding/json"
	"net/http"

	"ai-japanese-learning/internal/service"
)

type aiConfigRequest struct {
	Provider     string `json:"provider"`
	ProviderName string `json:"provider_name"`
	BaseURL      string `json:"base_url"`
	APIKey       string `json:"api_key"`
	Model        string `json:"model"`
	APIVersion   string `json:"api_version"`
}

func (r *Router) handleAIProviders(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"items": service.AIProviderDefinitions(),
	})
}

func (r *Router) handleAIConfig(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, r.aiService.CurrentStatus())
}

func (r *Router) handleAIConfigUpdate(w http.ResponseWriter, req *http.Request) {
	cfg, ok := r.decodeAIConfig(w, req)
	if !ok {
		return
	}
	status, err := r.aiService.ConfigureProvider(cfg)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  status,
		"message": "AI provider configured for current server process",
	})
}

func (r *Router) handleAIModels(w http.ResponseWriter, req *http.Request) {
	cfg, ok := r.decodeAIConfig(w, req)
	if !ok {
		return
	}
	models, status, err := r.aiService.ListProviderModels(req.Context(), cfg)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error":  err.Error(),
			"status": status,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":  models,
		"status": status,
	})
}

func (r *Router) handleAICheck(w http.ResponseWriter, req *http.Request) {
	cfg, ok := r.decodeAIConfig(w, req)
	if !ok {
		return
	}
	status, err := r.aiService.CheckProvider(req.Context(), cfg)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"ok":     false,
			"error":  err.Error(),
			"status": status,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"status":  status,
		"message": "AI provider connection check passed",
	})
}

func (r *Router) decodeAIConfig(w http.ResponseWriter, req *http.Request) (service.AIProviderConfig, bool) {
	var input aiConfigRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return service.AIProviderConfig{}, false
	}
	return service.AIProviderConfig{
		Provider:     input.Provider,
		ProviderName: input.ProviderName,
		BaseURL:      input.BaseURL,
		APIKey:       input.APIKey,
		Model:        input.Model,
		APIVersion:   input.APIVersion,
	}, true
}
