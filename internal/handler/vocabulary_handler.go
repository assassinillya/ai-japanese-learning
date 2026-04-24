package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type createVocabularyRequest struct {
	DictionaryEntryID  int64  `json:"dictionary_entry_id"`
	ArticleID          *int64 `json:"article_id"`
	SourceSentenceID   *int64 `json:"source_sentence_id"`
	SelectedText       string `json:"selected_text"`
	SourceSentenceText string `json:"source_sentence_text"`
}

func (r *Router) handleCreateVocabulary(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var input createVocabularyRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	item, created, err := r.vocabularyService.Add(
		req.Context(),
		user.ID,
		input.DictionaryEntryID,
		input.ArticleID,
		input.SourceSentenceID,
		input.SelectedText,
		input.SourceSentenceText,
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, map[string]any{
		"item":    item,
		"created": created,
	})
}

func (r *Router) handleVocabularyCheck(w http.ResponseWriter, req *http.Request) {
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

	item, added, err := r.vocabularyService.Check(req.Context(), user.ID, entryID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"added": added,
		"item":  item,
	})
}
