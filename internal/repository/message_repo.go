package repository

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ak-softwaredev/akrobat-msg/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrMessageNotFound = errors.New("message not found")

type pgMessagesRepo struct {
	pool *pgxpool.Pool
}

func NewMessagesRepository(pool *pgxpool.Pool) *pgMessagesRepo {
	return &pgMessagesRepo{
		pool: pool,
	}
}

func (r *pgMessagesRepo) GetChatMessages(
	ctx context.Context,
	chatID int64,
	beforeID int64,
	limit int,
) ([]model.Message, error) {
	switch {
	case limit < 0:
		limit = 50
	case limit > 100:
		limit = 100
	}

	messages := make([]model.Message, 0, limit)

	rows, err := r.pool.Query(ctx,
		`SELECT
			id,
			chat_id,
			sender_id,
			content,
			created_at,
			updated_at
		FROM messages
		WHERE chat_id = $1
			AND ($2 = 0 OR id < $2)
		ORDER BY id DESC
		LIMIT $3`,
		chatID, beforeID, limit,
	)
	if err != nil {
		log.Printf("[ERROR] repository get messages: %v", err)
		return nil, fmt.Errorf("get messages: %w", err)
	}

	var m model.Message
	_, err = pgx.ForEachRow(
		rows,
		[]any{
			&m.ID,
			&m.ChatID,
			&m.SenderID,
			&m.Content,
			&m.CreatedAt,
			&m.UpdatedAt,
		},
		func() error {
			messages = append(messages, m)
			return nil
		},
	)
	if err != nil {
		log.Printf("[ERROR] repository get messages: %v", err)
		return nil, fmt.Errorf("get messages: %w", err)
	}

	return messages, nil
}

func (r *pgMessagesRepo) CreateMessage(
	ctx context.Context,
	chatID int64,
	userID int64,
	content string,
) (*model.Message, error) {
	var m model.Message

	err := r.pool.QueryRow(ctx,
		`INSERT INTO messages (chat_id, sender_id, content)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			chat_id,
			sender_id,
			content,
			created_at,
			updated_at`,
		chatID, userID, content,
	).Scan(
		&m.ID,
		&m.ChatID,
		&m.SenderID,
		&m.Content,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		log.Printf("[ERROR] repository create message: %v", err)
		return nil, fmt.Errorf("create message: %w", err)
	}
	return &m, nil
}

func (r *pgMessagesRepo) UpdateMessage(
	ctx context.Context,
	chatID int64,
	messageID int64,
	userID int64,
	content string,
) (*model.Message, error) {
	var m model.Message

	err := r.pool.QueryRow(ctx,
		`UPDATE messages
		SET content = $4, updated_at = NOW()
		WHERE chat_id = $1
			AND id = $2
			AND sender_id = $3
		RETURNING
			id,
			chat_id,
			sender_id,
			content,
			created_at,
			updated_at`,
		chatID, messageID, userID, content,
	).Scan(
		&m.ID,
		&m.ChatID,
		&m.SenderID,
		&m.Content,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("update message: %w", ErrMessageNotFound)
		}
		log.Printf("[ERROR] repository update message: %v", err)
		return nil, fmt.Errorf("update message: %w", err)
	}

	return &m, nil
}

func (r *pgMessagesRepo) DeleteMessage(
	ctx context.Context,
	chatID int64,
	messageID int64,
	userID int64,
) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM messages
		WHERE chat_id = $1
			AND id = $2
			AND sender_id = $3`,
		chatID, messageID, userID,
	)
	if err != nil {
		log.Printf("[ERROR] repository delete message: %v", err)
		return fmt.Errorf("delete message: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("delete message: %w", ErrMessageNotFound)
	}

	return nil
}
