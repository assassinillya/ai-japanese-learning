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

func dictionaryEntryIDFromPath(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseInt(parts[2], 10, 64)
}
