package handler

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (r *Router) handleReadingArticle(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	articleID, err := articleIDFromPath(req.URL.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid article id"})
		return
	}

	article, err := r.articleService.Get(req.Context(), user.ID, articleID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	sentences, err := r.articleService.ListSentences(req.Context(), user.ID, articleID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"article":   article,
		"sentences": sentences,
	})
}

func (r *Router) handleGenerateChallengeQuestions(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	articleID, err := articleIDFromPath(strings.TrimSuffix(req.URL.Path, "/challenge-questions"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid article id"})
		return
	}

	questions, err := r.challengeService.Generate(req.Context(), user.ID, articleID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": questions})
}

func (r *Router) handleListChallengeQuestions(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	articleID, err := articleIDFromPath(strings.TrimSuffix(req.URL.Path, "/challenge-questions"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid article id"})
		return
	}

	questions, err := r.challengeService.GetOrGenerate(req.Context(), user.ID, articleID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": questions})
}

func (r *Router) handleSubmitChallengeAnswer(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	questionID, err := articleIDFromPath(strings.TrimSuffix(req.URL.Path, "/answer"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid question id"})
		return
	}

	var input struct {
		SelectedOption string `json:"selected_option"`
	}
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := r.challengeService.SubmitAnswer(req.Context(), user.ID, questionID, input.SelectedOption)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}
