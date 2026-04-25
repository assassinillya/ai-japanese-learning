package app

import (
	"database/sql"

	"ai-japanese-learning/internal/config"
	"ai-japanese-learning/internal/db"
	"ai-japanese-learning/internal/handler"
	"ai-japanese-learning/internal/repository"
	"ai-japanese-learning/internal/service"
	"ai-japanese-learning/internal/web"
)

type App struct {
	db     *sql.DB
	router *handler.Router
}

func New(cfg *config.Config) (*App, error) {
	postgres, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	userRepo := repository.NewUserRepository(postgres)
	articleRepo := repository.NewArticleRepository(postgres)
	dictionaryRepo := repository.NewDictionaryRepository(postgres)
	vocabularyRepo := repository.NewVocabularyRepository(postgres)
	challengeRepo := repository.NewChallengeRepository(postgres)
	reviewRepo := repository.NewReviewRepository(postgres)
	aiRepo := repository.NewAIRepository(postgres)
	authService := service.NewAuthService(userRepo, cfg.TokenSecret)
	profileService := service.NewProfileService(userRepo)
	languageService := service.NewLanguageService()
	aiService := service.NewAIService(aiRepo)
	translationService := service.NewTranslationService(aiService)
	articleService := service.NewArticleService(articleRepo, languageService, translationService)
	dictionaryService := service.NewDictionaryService(dictionaryRepo, aiService)
	vocabularyService := service.NewVocabularyService(vocabularyRepo, dictionaryRepo, articleRepo)
	challengeService := service.NewChallengeService(articleRepo, dictionaryRepo, challengeRepo, dictionaryService, aiService)
	reviewService := service.NewReviewService(vocabularyRepo, dictionaryRepo, reviewRepo, aiService)

	router := handler.NewRouter(
		authService,
		profileService,
		articleService,
		dictionaryService,
		vocabularyService,
		challengeService,
		reviewService,
		web.NewStaticServer(),
	)

	return &App{
		db:     postgres,
		router: router,
	}, nil
}

func (a *App) Router() *handler.Router {
	return a.router
}

func (a *App) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}
