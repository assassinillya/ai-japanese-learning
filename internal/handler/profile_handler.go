package handler

import (
	"encoding/json"
	"net/http"

	"ai-japanese-learning/internal/model"
)

type updateJLPTRequest struct {
	JLPTLevel model.JLPTLevel `json:"jlpt_level"`
}

func (r *Router) handleProfile(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	profile, err := r.profileService.GetProfile(req.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

func (r *Router) handleUpdateJLPTLevel(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var input updateJLPTRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := r.profileService.UpdateJLPTLevel(req.Context(), user.ID, input.JLPTLevel); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	profile, err := r.profileService.GetProfile(req.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

func (r *Router) handleCompleteOnboarding(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if err := r.profileService.CompleteOnboarding(req.Context(), user.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	profile, err := r.profileService.GetProfile(req.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, profile)
}
