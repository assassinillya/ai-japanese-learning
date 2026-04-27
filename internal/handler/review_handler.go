package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type reviewAnswerRequest struct {
	UserVocabularyID int64  `json:"user_vocabulary_id"`
	ReviewQuestionID int64  `json:"review_question_id"`
	SelectedOption   string `json:"selected_option"`
}

func (r *Router) handleReviewDue(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	limit := 20
	if raw := req.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil {
			limit = parsed
		}
	}
	extra := req.URL.Query().Get("extra") == "1" || req.URL.Query().Get("extra") == "true"
	var items any
	if extra {
		items, err = r.reviewService.Extra(req.Context(), user.ID, limit)
	} else {
		items, err = r.reviewService.Due(req.Context(), user.ID, limit)
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (r *Router) handleReviewQuestions(w http.ResponseWriter, req *http.Request) {
	r.handleReviewDue(w, req)
}

func (r *Router) handleReviewQuestion(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	entryID, err := strconv.ParseInt(req.URL.Query().Get("dictionary_entry_id"), 10, 64)
	if err != nil || entryID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid dictionary_entry_id"})
		return
	}
	entry, err := r.dictionaryService.GetByID(req.Context(), entryID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	question, err := r.reviewService.NextQuestion(req.Context(), user.ID, *entry)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"question": question})
}

func (r *Router) handleReviewPrewarm(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	created, err := r.reviewService.EnsureQuestionsForUser(req.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"created": created})
}

func (r *Router) handleReviewRecords(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	limit := 20
	if raw := req.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil {
			limit = parsed
		}
	}
	items, err := r.reviewService.Records(req.Context(), user.ID, limit)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (r *Router) handleReviewAnswer(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var input reviewAnswerRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := r.reviewService.SubmitAnswer(
		req.Context(),
		user.ID,
		input.UserVocabularyID,
		input.ReviewQuestionID,
		input.SelectedOption,
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}
