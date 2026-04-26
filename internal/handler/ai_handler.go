package handler

import (
	"encoding/json"
	"net/http"

	"ai-japanese-learning/internal/repository"
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
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if r.userAIConfigRepo != nil {
		cfg, err := r.userAIConfigRepo.Get(req.Context(), user.ID)
		if err == nil {
			writeJSON(w, http.StatusOK, r.aiService.StatusForConfig(aiConfigFromRepository(cfg)))
			return
		}
		if err != repository.ErrUserAIConfigNotFound {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}
	writeJSON(w, http.StatusOK, r.aiService.CurrentStatus())
}

func (r *Router) handleAIConfigUpdate(w http.ResponseWriter, req *http.Request) {
	cfg, ok := r.decodeAIConfig(w, req)
	if !ok {
		return
	}
	cfg = r.mergeStoredAIConfig(req, cfg)
	status := r.aiService.StatusForConfig(cfg)
	if r.userAIConfigRepo != nil {
		user, err := currentUser(req.Context())
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if err := r.userAIConfigRepo.Upsert(req.Context(), user.ID, aiConfigToRepository(cfg)); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  status,
		"message": "AI provider saved for current user",
	})
}

func aiConfigFromRepository(cfg repository.UserAIConfig) service.AIProviderConfig {
	return service.AIProviderConfig{
		Provider:     cfg.Provider,
		ProviderName: cfg.ProviderName,
		BaseURL:      cfg.BaseURL,
		APIKey:       cfg.APIKey,
		Model:        cfg.Model,
		APIVersion:   cfg.APIVersion,
	}
}

func aiConfigToRepository(cfg service.AIProviderConfig) repository.UserAIConfig {
	cfg = service.NormalizeAIProviderConfig(cfg)
	return repository.UserAIConfig{
		Provider:     cfg.Provider,
		ProviderName: cfg.ProviderName,
		BaseURL:      cfg.BaseURL,
		APIKey:       cfg.APIKey,
		Model:        cfg.Model,
		APIVersion:   cfg.APIVersion,
	}
}

func (r *Router) handleAIModels(w http.ResponseWriter, req *http.Request) {
	cfg, ok := r.decodeAIConfig(w, req)
	if !ok {
		return
	}
	cfg = r.mergeStoredAIConfig(req, cfg)
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
	cfg = r.mergeStoredAIConfig(req, cfg)
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

func (r *Router) mergeStoredAIConfig(req *http.Request, cfg service.AIProviderConfig) service.AIProviderConfig {
	if cfg.APIKey != "" || r.userAIConfigRepo == nil {
		return cfg
	}
	user, err := currentUser(req.Context())
	if err != nil {
		return cfg
	}
	stored, err := r.userAIConfigRepo.Get(req.Context(), user.ID)
	if err != nil {
		return cfg
	}
	cfg.APIKey = stored.APIKey
	return cfg
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
