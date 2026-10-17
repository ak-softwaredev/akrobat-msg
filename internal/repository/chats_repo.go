package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ak-softwaredev/akrobat-msg/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrChatNotFound = errors.New("chat not found")
var ErrChatMemberNotFound = errors.New("chat member not found")

type ChatsRepository interface {
	CreateChat(ctx context.Context, userID int64, name string) (*model.Chat, error)
	GetUserChats(ctx context.Context, userID int64) ([]model.Chat, error)
	GetChatInfo(ctx context.Context, chatID int64) (*model.Chat, error)
	UpdateChatDescription(ctx context.Context, chatID int64, description string) (*model.Chat, error)
	UpdateChatName(ctx context.Context, chatID int64, name string) (*model.Chat, error)
	DeleteChat(ctx context.Context, chatID int64) error
	AddMember(ctx context.Context, chatID int64, userID int64) error
	RemoveMember(ctx context.Context, chatID int64, userID int64) error
}

type pgChatsRepo struct {
	pool *pgxpool.Pool
}

func NewChatsRepository(pool *pgxpool.Pool) ChatsRepository {
	return &pgChatsRepo{
		pool: pool,
	}
}

func (r *pgChatsRepo) CreateChat(
	ctx context.Context,
	userID int64,
	name string,
) (*model.Chat, error) {
	var c model.Chat

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO chats (name, created_by)
		VALUES ($1, $2)
		RETURNING id, name, description, created_by, created_at, updated_at`,
		name, userID,
	).Scan(
		&c.ID,
		&c.Name,
		&c.Description,
		&c.CreatedBy,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO chat_members (chat_id, user_id, role)
		VALUES ($1, $2, $3)`,
		c.ID,
		userID,
		"owner",
	)
	if err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}

	return &c, nil
}

func (r *pgChatsRepo) GetUserChats(
	ctx context.Context,
	userID int64,
) ([]model.Chat, error) {
	chats := make([]model.Chat, 0)

	rows, err := r.pool.Query(ctx,
		`SELECT 
			c.id, 
			c.name, 
			c.description, 
			c.created_by, 
			c.created_at, 
			c.updated_at
		FROM chats AS c
		JOIN chat_members AS cm ON cm.chat_id = c.id
		WHERE cm.user_id = $1
		ORDER BY c.updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get chats: %w", err)
	}

	var c model.Chat

	_, err = pgx.ForEachRow(
		rows,
		[]any{
			&c.ID,
			&c.Name,
			&c.Description,
			&c.CreatedBy,
			&c.CreatedAt,
			&c.UpdatedAt,
		},
		func() error {
			chats = append(chats, c)
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("get chats: %w", err)
	}

	return chats, nil
}

func (r *pgChatsRepo) GetChatInfo(
	ctx context.Context,
	chatID int64,
) (*model.Chat, error) {
	var c model.Chat

	err := r.pool.QueryRow(ctx,
		`SELECT
			id,
			name,
			description,
			created_by,
			created_at,
			updated_at
		FROM chats
		WHERE id = $1`,
		chatID,
	).Scan(
		&c.ID,
		&c.Name,
		&c.Description,
		&c.CreatedBy,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get chat info: %w", ErrChatNotFound)
		}
		return nil, fmt.Errorf("get chat info: %w", err)
	}

	return &c, nil
}

func (r *pgChatsRepo) UpdateChatDescription(
	ctx context.Context,
	chatID int64,
	description string,
) (*model.Chat, error) {
	var c model.Chat
	err := r.pool.QueryRow(ctx,
		`UPDATE chats
		SET description = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING
			id,
			name,
			description,
			created_by,
			created_at,
			updated_at`,
		chatID, description,
	).Scan(
		&c.ID,
		&c.Name,
		&c.Description,
		&c.CreatedBy,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("update chat: %w", ErrChatNotFound)
		}
		return nil, fmt.Errorf("update chat: %w", err)
	}
	return &c, nil
}

func (r *pgChatsRepo) UpdateChatName(
	ctx context.Context,
	chatID int64,
	name string,
) (*model.Chat, error) {
	var c model.Chat
	err := r.pool.QueryRow(ctx,
		`UPDATE chats
		SET name = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING
			id,
			name,
			name,
			created_by,
			created_at,
			updated_at`,
		chatID, name,
	).Scan(
		&c.ID,
		&c.Name,
		&c.Description,
		&c.CreatedBy,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("update chat: %w", ErrChatNotFound)
		}
		return nil, fmt.Errorf("update chat: %w", err)
	}
	return &c, nil
}

func (r *pgChatsRepo) DeleteChat(
	ctx context.Context,
	chatID int64,
) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM chats
		WHERE id = $1`,
		chatID,
	)
	if err != nil {
		return fmt.Errorf("delete chat: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("delete chat: %w", ErrChatNotFound)
	}

	return nil
}

func (r *pgChatsRepo) AddMember(
	ctx context.Context,
	chatID int64,
	userID int64,
) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO chat_members (chat_id, user_id)
		VALUES ($1, $2)`,
		chatID, userID,
	)
	if err != nil {
		return fmt.Errorf("add member: %w", err)
	}

	return nil
}

func (r *pgChatsRepo) RemoveMember(
	ctx context.Context,
	chatID int64,
	userID int64,
) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM chat_members
		WHERE chat_id = $1
			AND user_id = $2`,
		chatID, userID,
	)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("remove member: %w", ErrChatMemberNotFound)
	}

	return nil
}
