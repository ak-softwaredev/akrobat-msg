package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ak-softwaredev/akrobat-msg/internal/model"
	"github.com/ak-softwaredev/akrobat-msg/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthSuccessResult struct {
	Token   string         `json:"token"`
	Session *model.Session `json:"session"`
}

type AuthService interface {
	Register(ctx context.Context, username, password string) (*AuthSuccessResult, error)
	Login(ctx context.Context, username, password string) (*AuthSuccessResult, error)
	Logout(ctx context.Context, token string) error
	Authenticate(ctx context.Context, token string) (int64, error)
}

type authService struct {
	users    repository.UsersRepository
	sessions repository.SessionRepository
}

func NewAuthService(
	users repository.UsersRepository,
	sessions repository.SessionRepository,
) AuthService {
	return &authService{
		users:    users,
		sessions: sessions,
	}
}

func (s *authService) Register(
	ctx context.Context,
	username, password string,
) (*AuthSuccessResult, error) {
	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, fmt.Errorf("auth register: %w", err)
	}

	user, err := s.users.Register(
		ctx,
		username,
		string(passwordHash),
	)
	if err != nil {
		return nil, fmt.Errorf("auth register: %w", err)
	}

	token, err := generateToken(64)
	if err != nil {
		return nil, fmt.Errorf("auth register: %w", err)
	}
	tokenHash := hashToken(token)

	session, err := s.sessions.Create(
		ctx,
		repository.CreateSessionParams{
			TokenHash: tokenHash,
			UserID:    user.ID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("auth register: %w", err)
	}

	return &AuthSuccessResult{
		Token:   token,
		Session: session,
	}, nil
}

func (s *authService) Login(
	ctx context.Context,
	username, password string,
) (*AuthSuccessResult, error) {
	user, err := s.users.GetWithCredentialsByUsername(
		ctx,
		username,
	)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("auth login: %w", err)
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := generateToken(64)
	if err != nil {
		return nil, fmt.Errorf("auth login: %w", err)
	}
	tokenHash := hashToken(token)

	session, err := s.sessions.Create(
		ctx,
		repository.CreateSessionParams{
			TokenHash: string(tokenHash),
			UserID:    user.ID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("auth login: %w", err)
	}

	return &AuthSuccessResult{
		Token:   token,
		Session: session,
	}, nil
}

func (s *authService) Logout(
	ctx context.Context,
	token string,
) error {
	hash := hashToken(token)
	err := s.sessions.DeleteByHash(
		ctx,
		hash,
	)
	if err != nil {
		return fmt.Errorf("auth logout: %w", err)
	}
	return nil
}

func (s *authService) Authenticate(
	ctx context.Context,
	token string,
) (int64, error) {
	tokenHash := hashToken(token)
	session, err := s.sessions.GetByHash(
		ctx,
		tokenHash,
	)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return -1, ErrInvalidCredentials
		}
		return -1, fmt.Errorf("auth authenticate: %w", err)
	}

	return session.UserID, nil
}

func generateToken(length int) (string, error) {
	buffer := make([]byte, length)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(buffer)
	return "auth_" + token, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
