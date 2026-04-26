package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"ai-japanese-learning/internal/model"
	"ai-japanese-learning/internal/service"
)

type contextKey string

const userContextKey contextKey = "current_user"

type Router struct {
	mux               *http.ServeMux
	authService       *service.AuthService
	profileService    *service.ProfileService
	articleService    *service.ArticleService
	dictionaryService *service.DictionaryService
	vocabularyService *service.VocabularyService
	challengeService  *service.ChallengeService
	reviewService     *service.ReviewService
	statsService      *service.StatsService
}

type staticServer interface {
	Handler() http.Handler
}

func NewRouter(
	authService *service.AuthService,
	profileService *service.ProfileService,
	articleService *service.ArticleService,
	dictionaryService *service.DictionaryService,
	vocabularyService *service.VocabularyService,
	challengeService *service.ChallengeService,
	reviewService *service.ReviewService,
	statsService *service.StatsService,
	static staticServer,
) *Router {
	r := &Router{
		mux:               http.NewServeMux(),
		authService:       authService,
		profileService:    profileService,
		articleService:    articleService,
		dictionaryService: dictionaryService,
		vocabularyService: vocabularyService,
		challengeService:  challengeService,
		reviewService:     reviewService,
		statsService:      statsService,
	}

	r.mux.Handle("/assets/", http.StripPrefix("/assets/", static.Handler()))
	r.mux.Handle("/", static.Handler())

	r.mux.HandleFunc("POST /api/auth/register", r.handleRegister)
	r.mux.HandleFunc("GET /api/health", r.handleHealth)
	r.mux.HandleFunc("POST /api/auth/login", r.handleLogin)
	r.mux.HandleFunc("POST /api/auth/logout", r.withAuth(r.handleLogout))
	r.mux.HandleFunc("GET /api/auth/me", r.withAuth(r.handleMe))
	r.mux.HandleFunc("GET /api/profile", r.withAuth(r.handleProfile))
	r.mux.HandleFunc("PUT /api/profile/jlpt-level", r.withAuth(r.handleUpdateJLPTLevel))
	r.mux.HandleFunc("POST /api/profile/onboarding/complete", r.withAuth(r.handleCompleteOnboarding))
	r.mux.HandleFunc("GET /api/articles/library", r.withAuth(r.handleLibraryArticles))
	r.mux.HandleFunc("GET /api/articles", r.withAuth(r.handleMyArticles))
	r.mux.HandleFunc("POST /api/articles", r.withAuth(r.handleCreateArticle))
	r.mux.HandleFunc("POST /api/articles/upload", r.withAuth(r.handleCreateArticle))
	r.mux.HandleFunc("GET /api/articles/{id}", r.withAuth(r.handleArticleDetail))
	r.mux.HandleFunc("POST /api/articles/{id}/process", r.withAuth(r.handleArticleProcess))
	r.mux.HandleFunc("GET /api/articles/{id}/sentences", r.withAuth(r.handleArticleSentences))
	r.mux.HandleFunc("GET /api/reading/articles/{id}", r.withAuth(r.handleReadingArticle))
	r.mux.HandleFunc("POST /api/reading/articles/{id}/challenge-questions", r.withAuth(r.handleGenerateChallengeQuestions))
	r.mux.HandleFunc("GET /api/reading/articles/{id}/challenge-questions", r.withAuth(r.handleListChallengeQuestions))
	r.mux.HandleFunc("POST /api/reading/articles/{id}/post-quiz", r.withAuth(r.handleGeneratePostQuiz))
	r.mux.HandleFunc("GET /api/reading/articles/{id}/post-quiz", r.withAuth(r.handleListPostQuiz))
	r.mux.HandleFunc("GET /api/reading/articles/{id}/post-quiz/results", r.withAuth(r.handlePostQuizResults))
	r.mux.HandleFunc("POST /api/reading/questions/{id}/answer", r.withAuth(r.handleSubmitChallengeAnswer))
	r.mux.HandleFunc("GET /api/dictionary/search", r.withAuth(r.handleDictionarySearch))
	r.mux.HandleFunc("GET /api/dictionary/lookup", r.withAuth(r.handleDictionaryLookup))
	r.mux.HandleFunc("POST /api/dictionary/generate", r.withAuth(r.handleDictionaryGenerate))
	r.mux.HandleFunc("GET /api/dictionary/{id}", r.withAuth(r.handleDictionaryDetail))
	r.mux.HandleFunc("GET /api/vocabulary", r.withAuth(r.handleListVocabulary))
	r.mux.HandleFunc("POST /api/vocabulary", r.withAuth(r.handleCreateVocabulary))
	r.mux.HandleFunc("GET /api/vocabulary/check", r.withAuth(r.handleVocabularyCheck))
	r.mux.HandleFunc("GET /api/vocabulary/{id}", r.withAuth(r.handleVocabularyDetail))
	r.mux.HandleFunc("GET /api/vocabulary/{id}/context", r.withAuth(r.handleVocabularyContext))
	r.mux.HandleFunc("PUT /api/vocabulary/{id}/status", r.withAuth(r.handleUpdateVocabularyStatus))
	r.mux.HandleFunc("DELETE /api/vocabulary/{id}", r.withAuth(r.handleDeleteVocabulary))
	r.mux.HandleFunc("GET /api/review/due", r.withAuth(r.handleReviewDue))
	r.mux.HandleFunc("POST /api/review/questions", r.withAuth(r.handleReviewQuestions))
	r.mux.HandleFunc("POST /api/review/answer", r.withAuth(r.handleReviewAnswer))
	r.mux.HandleFunc("GET /api/review/records", r.withAuth(r.handleReviewRecords))
	r.mux.HandleFunc("GET /api/stats/learning", r.withAuth(r.handleLearningStats))

	return r
}

func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": "v1.2",
	})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	r.mux.ServeHTTP(w, req)
}

func (r *Router) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		token := extractBearerToken(req.Header.Get("Authorization"))
		user, err := r.authService.Authenticate(req.Context(), token)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		ctx := context.WithValue(req.Context(), userContextKey, user)
		next(w, req.WithContext(ctx))
	}
}

func currentUser(ctx context.Context) (*model.User, error) {
	user, ok := ctx.Value(userContextKey).(*model.User)
	if !ok || user == nil {
		return nil, errors.New("current user missing")
	}
	return user, nil
}

func extractBearerToken(header string) string {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
