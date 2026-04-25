package service

import (
	"context"
	"fmt"

	"ai-japanese-learning/internal/model"
	"ai-japanese-learning/internal/repository"
)

type ProfileService struct {
	userRepo *repository.UserRepository
}

func NewProfileService(userRepo *repository.UserRepository) *ProfileService {
	return &ProfileService{userRepo: userRepo}
}

func (s *ProfileService) GetProfile(ctx context.Context, userID int64) (*model.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

func (s *ProfileService) UpdateJLPTLevel(ctx context.Context, userID int64, level model.JLPTLevel) error {
	if !model.IsValidJLPT(level) {
		return fmt.Errorf("invalid jlpt level")
	}
	return s.userRepo.UpdateJLPTLevel(ctx, userID, level)
}

func (s *ProfileService) CompleteOnboarding(ctx context.Context, userID int64) error {
	return s.userRepo.CompleteOnboarding(ctx, userID)
}
