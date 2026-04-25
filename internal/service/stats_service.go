package service

import (
	"context"

	"ai-japanese-learning/internal/model"
	"ai-japanese-learning/internal/repository"
)

type StatsService struct {
	statsRepo *repository.StatsRepository
}

func NewStatsService(statsRepo *repository.StatsRepository) *StatsService {
	return &StatsService{statsRepo: statsRepo}
}

func (s *StatsService) LearningStats(ctx context.Context, userID int64) (*model.LearningStats, error) {
	return s.statsRepo.GetLearningStats(ctx, userID)
}
