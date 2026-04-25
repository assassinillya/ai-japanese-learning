package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"ai-japanese-learning/internal/model"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, email, username, passwordHash string, level model.JLPTLevel) (*model.User, error) {
	query := `
		INSERT INTO users (email, username, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, email, username, created_at, updated_at
	`

	var user model.User
	if err := r.db.QueryRowContext(ctx, query, email, username, passwordHash).
		Scan(&user.ID, &user.Email, &user.Username, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	profileQuery := `
		INSERT INTO user_profiles (user_id, jlpt_level, onboarding_completed)
		VALUES ($1, $2, FALSE)
	`
	if _, err := r.db.ExecContext(ctx, profileQuery, user.ID, level); err != nil {
		return nil, fmt.Errorf("insert profile: %w", err)
	}

	user.JLPTLevel = level
	user.OnboardingCompleted = false
	return &user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, string, error) {
	query := `
		SELECT u.id, u.email, u.username, u.password_hash, u.created_at, u.updated_at,
		       COALESCE(p.jlpt_level, 'N5'), COALESCE(p.onboarding_completed, FALSE), p.last_login_at
		FROM users u
		LEFT JOIN user_profiles p ON p.user_id = u.id
		WHERE u.email = $1
	`

	var user model.User
	var passwordHash string
	if err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&passwordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.JLPTLevel,
		&user.OnboardingCompleted,
		&user.LastLoginAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", ErrUserNotFound
		}
		return nil, "", fmt.Errorf("find user by email: %w", err)
	}

	return &user, passwordHash, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
	query := `
		SELECT u.id, u.email, u.username, u.created_at, u.updated_at,
		       COALESCE(p.jlpt_level, 'N5'), COALESCE(p.onboarding_completed, FALSE), p.last_login_at
		FROM users u
		LEFT JOIN user_profiles p ON p.user_id = u.id
		WHERE u.id = $1
	`

	var user model.User
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.JLPTLevel,
		&user.OnboardingCompleted,
		&user.LastLoginAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) UpsertSession(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO auth_sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	if _, err := r.db.ExecContext(ctx, query, userID, tokenHash, expiresAt); err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (r *UserRepository) FindUserByTokenHash(ctx context.Context, tokenHash string) (*model.User, error) {
	query := `
		SELECT u.id, u.email, u.username, u.created_at, u.updated_at,
		       COALESCE(p.jlpt_level, 'N5'), COALESCE(p.onboarding_completed, FALSE), p.last_login_at
		FROM auth_sessions s
		INNER JOIN users u ON u.id = s.user_id
		LEFT JOIN user_profiles p ON p.user_id = u.id
		WHERE s.token_hash = $1 AND s.expires_at > NOW()
		ORDER BY s.id DESC
		LIMIT 1
	`

	var user model.User
	if err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.JLPTLevel,
		&user.OnboardingCompleted,
		&user.LastLoginAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by token hash: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) DeleteSession(ctx context.Context, tokenHash string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE token_hash = $1`, tokenHash); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateJLPTLevel(ctx context.Context, userID int64, level model.JLPTLevel) error {
	query := `
		UPDATE user_profiles
		SET jlpt_level = $2, onboarding_completed = TRUE, updated_at = NOW()
		WHERE user_id = $1
	`
	result, err := r.db.ExecContext(ctx, query, userID, level)
	if err != nil {
		return fmt.Errorf("update jlpt level: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update jlpt level rows: %w", err)
	}
	if rows == 0 {
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO user_profiles (user_id, jlpt_level, onboarding_completed)
			VALUES ($1, $2, TRUE)
		`, userID, level)
		if err != nil {
			return fmt.Errorf("insert profile during jlpt update: %w", err)
		}
	}
	return nil
}

func (r *UserRepository) TouchLastLogin(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_profiles
		SET last_login_at = NOW(), updated_at = NOW()
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("touch last login: %w", err)
	}
	return nil
}

func (r *UserRepository) CompleteOnboarding(ctx context.Context, userID int64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE user_profiles
		SET onboarding_completed = TRUE,
		    updated_at = NOW()
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("complete onboarding: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("complete onboarding rows: %w", err)
	}
	if rows == 0 {
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO user_profiles (user_id, jlpt_level, onboarding_completed)
			VALUES ($1, 'N5', TRUE)
		`, userID)
		if err != nil {
			return fmt.Errorf("insert profile during onboarding completion: %w", err)
		}
	}
	return nil
}
