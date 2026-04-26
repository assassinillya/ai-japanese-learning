package app

import (
	"context"
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
	if err := dictionaryRepo.EnsureExampleTable(context.Background()); err != nil {
		_ = postgres.Close()
		return nil, err
	}
	vocabularyRepo := repository.NewVocabularyRepository(postgres)
	challengeRepo := repository.NewChallengeRepository(postgres)
	reviewRepo := repository.NewReviewRepository(postgres)
	aiRepo := repository.NewAIRepository(postgres)
	userAIConfigRepo := repository.NewUserAIConfigRepository(postgres)
	if err := userAIConfigRepo.EnsureTable(context.Background()); err != nil {
		_ = postgres.Close()
		return nil, err
	}
	statsRepo := repository.NewStatsRepository(postgres)
	authService := service.NewAuthService(userRepo, cfg.TokenSecret)
	profileService := service.NewProfileService(userRepo)
	languageService := service.NewLanguageService()
	aiService := service.NewConfiguredAIService(aiRepo, service.AIProviderConfig{
		Provider:     cfg.AI.Provider,
		ProviderName: cfg.AI.ProviderName,
		BaseURL:      cfg.AI.BaseURL,
		APIKey:       cfg.AI.APIKey,
		Model:        cfg.AI.Model,
		APIVersion:   cfg.AI.APIVersion,
	})
	translationService := service.NewTranslationService(aiService)
	articleService := service.NewArticleService(articleRepo, languageService, translationService)
	dictionaryService := service.NewDictionaryService(dictionaryRepo, aiService)
	vocabularyService := service.NewVocabularyService(vocabularyRepo, dictionaryRepo, articleRepo)
	challengeService := service.NewChallengeService(articleRepo, dictionaryRepo, challengeRepo, dictionaryService, aiService)
	reviewService := service.NewReviewService(vocabularyRepo, dictionaryRepo, reviewRepo, aiService)
	statsService := service.NewStatsService(statsRepo)

	router := handler.NewRouter(
		authService,
		profileService,
		articleService,
		dictionaryService,
		vocabularyService,
		challengeService,
		reviewService,
		statsService,
		aiService,
		userAIConfigRepo,
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
