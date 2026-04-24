package handler

import "net/http"

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
