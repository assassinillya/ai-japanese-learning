package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"ai-japanese-learning/internal/model"
	"ai-japanese-learning/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

type AuthResult struct {
	User        *model.User `json:"user"`
	AccessToken string      `json:"access_token"`
	ExpiresAt   time.Time   `json:"expires_at"`
}

type AuthService struct {
	userRepo    *repository.UserRepository
	tokenSecret string
}

func NewAuthService(userRepo *repository.UserRepository, tokenSecret string) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		tokenSecret: tokenSecret,
	}
}

func (s *AuthService) Register(ctx context.Context, email, username, password string, level model.JLPTLevel) (*AuthResult, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.userRepo.CreateUser(ctx, email, username, string(passwordHash), level)
	if err != nil {
		return nil, err
	}

	return s.createSession(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	user, passwordHash, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := s.userRepo.TouchLastLogin(ctx, user.ID); err != nil {
		return nil, err
	}

	return s.createSession(ctx, user)
}

func (s *AuthService) Authenticate(ctx context.Context, token string) (*model.User, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}

	user, err := s.userRepo.FindUserByTokenHash(ctx, hashToken(token))
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.userRepo.DeleteSession(ctx, hashToken(token))
}

func (s *AuthService) createSession(ctx context.Context, user *model.User) (*AuthResult, error) {
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := s.userRepo.UpsertSession(ctx, user.ID, hashToken(token), expiresAt); err != nil {
		return nil, err
	}

	return &AuthResult{
		User:        user,
		AccessToken: token,
		ExpiresAt:   expiresAt,
	}, nil
}

func generateToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum[:])
}
