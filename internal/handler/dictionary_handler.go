package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type dictionaryGenerateRequest struct {
	Text string `json:"text"`
}

type dictionaryExampleRequest struct {
	DictionaryEntryID int64 `json:"dictionary_entry_id"`
}

func (r *Router) handleDictionarySearch(w http.ResponseWriter, req *http.Request) {
	text := req.URL.Query().Get("text")
	entry, found, err := r.dictionaryService.Lookup(req.Context(), text)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"found": found,
		"entry": entry,
	})
}

func (r *Router) handleDictionaryLookup(w http.ResponseWriter, req *http.Request) {
	text := req.URL.Query().Get("text")
	entry, generated, err := r.dictionaryService.LookupOrGenerate(req.Context(), text)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"entry":     entry,
		"generated": generated,
	})
}

func (r *Router) handleDictionaryGenerate(w http.ResponseWriter, req *http.Request) {
	var input dictionaryGenerateRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	entry, generated, err := r.dictionaryService.LookupOrGenerate(req.Context(), input.Text)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"entry":     entry,
		"generated": generated,
	})
}

func (r *Router) handleDictionaryDetail(w http.ResponseWriter, req *http.Request) {
	entryID, err := dictionaryEntryIDFromPath(req.URL.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid dictionary entry id"})
		return
	}

	entry, err := r.dictionaryService.GetByID(req.Context(), entryID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, entry)
}

func (r *Router) handleDictionaryExamples(w http.ResponseWriter, req *http.Request) {
	entryID, err := dictionaryEntryIDFromPath(strings.TrimSuffix(req.URL.Path, "/examples"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid dictionary entry id"})
		return
	}
	items, err := r.dictionaryService.ListExamples(req.Context(), entryID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (r *Router) handleDictionaryExampleGenerate(w http.ResponseWriter, req *http.Request) {
	var input dictionaryExampleRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	example, err := r.dictionaryService.GenerateExample(req.Context(), input.DictionaryEntryID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, example)
}

func (r *Router) handleDictionaryExampleDelete(w http.ResponseWriter, req *http.Request) {
	exampleID, err := dictionaryExampleIDFromPath(req.URL.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid dictionary example id"})
		return
	}
	if err := r.dictionaryService.DeleteExample(req.Context(), exampleID); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func dictionaryEntryIDFromPath(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseInt(parts[2], 10, 64)
}

func dictionaryExampleIDFromPath(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 4 {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseInt(parts[len(parts)-1], 10, 64)
}
