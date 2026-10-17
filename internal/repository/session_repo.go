package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ak-softwaredev/akrobat-msg/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSessionNotFound = errors.New("session not found")

type CreateSessionParams struct {
	TokenHash string
	UserID    int64
	ExpiresAt time.Time
}

type SessionRepository interface {
	Create(ctx context.Context, p CreateSessionParams) (*model.Session, error)
	GetByHash(ctx context.Context, tokenHash string) (*model.Session, error)
	GetByUserID(ctx context.Context, id int64) ([]model.Session, error)
	DeleteByHash(ctx context.Context, tokenHash string) error
	DeleteByUserID(ctx context.Context, id int64) error
	DeleteExpired(ctx context.Context) error
}

type pgSessionRepo struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) SessionRepository {
	return &pgSessionRepo{
		pool: pool,
	}
}

func (r *pgSessionRepo) Create(
	ctx context.Context,
	p CreateSessionParams,
) (*model.Session, error) {
	var s model.Session

	err := r.pool.QueryRow(ctx,
		`INSERT INTO sessions (token_hash, user_id, expires_at)
		VALUES ($1, $2, $3)
		RETURNING token_hash, user_id, expires_at`,
		p.TokenHash, p.UserID, p.ExpiresAt,
	).Scan(&s.TokenHash, &s.UserID, &s.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("session create: %w", err)
	}

	return &s, nil
}

func (r *pgSessionRepo) GetByHash(
	ctx context.Context,
	tokenHash string,
) (*model.Session, error) {
	var s model.Session

	err := r.pool.QueryRow(ctx,
		`SELECT token_hash, user_id, expires_at FROM sessions
		WHERE token_hash = $1
			AND expires_at > NOW()`,
		tokenHash,
	).Scan(&s.TokenHash, &s.UserID, &s.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("session get: %w", ErrSessionNotFound)
		}
		return nil, fmt.Errorf("session get: %w", err)
	}

	return &s, nil
}

func (r *pgSessionRepo) GetByUserID(
	ctx context.Context,
	id int64,
) ([]model.Session, error) {
	sessions := make([]model.Session, 0)

	rows, err := r.pool.Query(ctx,
		`SELECT token_hash, user_id, expires_at FROM sessions
		WHERE user_id = $1
			AND expires_at > NOW()`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("session get: %w", err)
	}

	var s model.Session

	_, err = pgx.ForEachRow(
		rows,
		[]any{
			&s.TokenHash,
			&s.UserID,
			&s.ExpiresAt,
		},
		func() error {
			sessions = append(sessions, s)
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("session get: %w", err)
	}

	return sessions, nil
}

func (r *pgSessionRepo) DeleteByHash(
	ctx context.Context,
	tokenHash string,
) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM sessions
		WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("session delete: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("session delete: %w", ErrSessionNotFound)
	}

	return nil
}

func (r *pgSessionRepo) DeleteByUserID(
	ctx context.Context,
	id int64,
) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM sessions
		WHERE user_id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("session delete: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("session delete: %w", ErrSessionNotFound)
	}

	return nil
}

func (r *pgSessionRepo) DeleteExpired(
	ctx context.Context,
) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM sessions
		WHERE expires_at <= NOW()`,
	)
	if err != nil {
		return fmt.Errorf("session delete: %w", err)
	}

	return nil
}
