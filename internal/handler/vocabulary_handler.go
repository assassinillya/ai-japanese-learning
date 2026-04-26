package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"ai-japanese-learning/internal/model"
	"ai-japanese-learning/internal/service"
)

type createVocabularyRequest struct {
	DictionaryEntryID  int64  `json:"dictionary_entry_id"`
	ArticleID          *int64 `json:"article_id"`
	SourceSentenceID   *int64 `json:"source_sentence_id"`
	SelectedText       string `json:"selected_text"`
	SourceSentenceText string `json:"source_sentence_text"`
}

type updateVocabularyStatusRequest struct {
	Status model.VocabularyStatus `json:"status"`
}

type batchVocabularyStatusRequest struct {
	VocabularyIDs []int64                `json:"vocabulary_ids"`
	Status        model.VocabularyStatus `json:"status"`
}

type batchVocabularyDeleteRequest struct {
	VocabularyIDs []int64 `json:"vocabulary_ids"`
}

func (r *Router) handleListVocabulary(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	items, err := r.vocabularyService.List(req.Context(), user.ID, req.URL.Query().Get("status"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
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
	if created {
		go func(ctx context.Context, entryID int64) {
			entry, err := r.dictionaryService.GetByID(ctx, entryID)
			if err != nil {
				return
			}
			_, _ = r.reviewService.GetOrCreateQuestion(ctx, *entry)
		}(service.BackgroundContextWithAIProvider(req.Context()), input.DictionaryEntryID)
	}
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

func (r *Router) handleVocabularyDetail(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	vocabularyID, err := vocabularyIDFromPath(req.URL.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid vocabulary id"})
		return
	}

	item, err := r.vocabularyService.GetDetail(req.Context(), user.ID, vocabularyID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (r *Router) handleVocabularyContext(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	vocabularyID, err := vocabularyIDFromPath(strings.TrimSuffix(req.URL.Path, "/context"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid vocabulary id"})
		return
	}

	item, err := r.vocabularyService.GetDetail(req.Context(), user.ID, vocabularyID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"selected_text":      item.Item.SelectedText,
		"example_sentence":   item.ExampleSentence,
		"article_id":         item.Item.ArticleID,
		"article_title":      item.ArticleTitle,
		"source_sentence_id": item.Item.SourceSentenceID,
	})
}

func (r *Router) handleUpdateVocabularyStatus(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	vocabularyID, err := vocabularyIDFromPath(strings.TrimSuffix(req.URL.Path, "/status"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid vocabulary id"})
		return
	}

	var input updateVocabularyStatusRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	item, err := r.vocabularyService.UpdateStatus(req.Context(), user.ID, vocabularyID, input.Status)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (r *Router) handleBatchUpdateVocabularyStatus(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var input batchVocabularyStatusRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	updated, err := r.vocabularyService.UpdateStatusBatch(req.Context(), user.ID, input.VocabularyIDs, input.Status)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"updated": updated})
}

func (r *Router) handleDeleteVocabulary(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	vocabularyID, err := vocabularyIDFromPath(req.URL.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid vocabulary id"})
		return
	}

	if err := r.vocabularyService.Delete(req.Context(), user.ID, vocabularyID); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (r *Router) handleBatchDeleteVocabulary(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var input batchVocabularyDeleteRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	deleted, err := r.vocabularyService.DeleteBatch(req.Context(), user.ID, input.VocabularyIDs)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": deleted})
}

func vocabularyIDFromPath(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseInt(parts[len(parts)-1], 10, 64)
}
