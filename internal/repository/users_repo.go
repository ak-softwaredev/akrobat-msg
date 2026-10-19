package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ak-softwaredev/akrobat-msg/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUserExists = errors.New("user already exists")
var ErrInvalidUserField = errors.New("invalid user field")

type pgUsersRepo struct {
	pool *pgxpool.Pool
}

func NewUsersRepository(pool *pgxpool.Pool) *pgUsersRepo {
	return &pgUsersRepo{
		pool: pool,
	}
}

func (r *pgUsersRepo) Register(
	ctx context.Context,
	username string,
	passwordHash string,
) (*model.User, error) {
	var u model.User

	_, err := r.GetByUsername(ctx, username)
	if err == nil {
		return nil, ErrUserExists
	}
	if !errors.Is(err, ErrUserNotFound) {
		return nil, fmt.Errorf("register user (pre-check): %w", err)
	}

	err = r.pool.QueryRow(ctx,
		`INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING id, username, created_at, updated_at`,
		username, passwordHash,
	).Scan(&u.ID, &u.Username, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "users_username_key" {
			return nil, ErrUserExists
		}
		return nil, fmt.Errorf("register user: %w", err)
	}
	return &u, nil
}

func (r *pgUsersRepo) GetWithCredentialsByUsername(
	ctx context.Context,
	username string,
) (*model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, COALESCE(bio, ''), password_hash, created_at, updated_at
		FROM users
		WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.Bio, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("login user: %w", ErrUserNotFound)
		}
		return nil, fmt.Errorf("login user: %w", err)
	}
	return &u, nil
}

func (r *pgUsersRepo) GetByUsername(
	ctx context.Context,
	username string,
) (*model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, COALESCE(bio, ''), created_at, updated_at
		FROM users
		WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.Bio, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &u, nil
}

func (r *pgUsersRepo) GetByID(
	ctx context.Context,
	id int64,
) (*model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, COALESCE(bio, ''), created_at, updated_at
		FROM users
		WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Username, &u.Bio, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &u, nil
}

func (r *pgUsersRepo) UpdateUsername(
	ctx context.Context,
	id int64,
	username string,
) (*model.User, error) {
	var u model.User

	err := r.pool.QueryRow(ctx,
		`UPDATE users
		SET username = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, username, COALESCE(bio, ''), created_at, updated_at`,
		username, id,
	).Scan(&u.ID, &u.Username, &u.Bio, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("update user: %w", err)
	}

	return &u, nil
}

func (r *pgUsersRepo) UpdatePassword(
	ctx context.Context,
	id int64,
	username string,
) (*model.User, error) {
	var u model.User

	err := r.pool.QueryRow(ctx,
		`UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, username, COALESCE(bio, ''), created_at, updated_at`,
		username, id,
	).Scan(&u.ID, &u.Username, &u.Bio, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("update user: %w", err)
	}

	return &u, nil
}

func (r *pgUsersRepo) UpdateBio(
	ctx context.Context,
	id int64,
	bio string,
) (*model.User, error) {
	var u model.User

	err := r.pool.QueryRow(ctx,
		`UPDATE users
		SET bio = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, username, COALESCE(bio, ''), created_at, updated_at`,
		bio, id,
	).Scan(&u.ID, &u.Username, &u.Bio, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("update user: %w", err)
	}

	return &u, nil
}

func (r *pgUsersRepo) Delete(
	ctx context.Context,
	id int64,
) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM users WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}
