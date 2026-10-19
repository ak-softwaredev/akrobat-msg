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

var ErrChatNotFound = errors.New("chat not found")
var ErrChatMemberNotFound = errors.New("chat member not found")

type pgChatsRepo struct {
	pool *pgxpool.Pool
}

type chatMemberQueryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func NewChatsRepository(pool *pgxpool.Pool) *pgChatsRepo {
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
		log.Printf("[ERROR] repository create chat: %v", err)
		return nil, fmt.Errorf("create chat: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO chats (name, created_by)
		VALUES ($1, $2)
		RETURNING id, name, COALESCE(description, ''), created_by, created_at, updated_at`,
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
		log.Printf("[ERROR] repository create chat: %v", err)
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
		log.Printf("[ERROR] repository create chat: %v", err)
		return nil, fmt.Errorf("create chat: %w", err)
	}
	c.Members, err = getChatMembers(ctx, tx, c.ID)
	if err != nil {
		log.Printf("[ERROR] repository create chat: %v", err)
		return nil, fmt.Errorf("create chat: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("[ERROR] repository create chat: %v", err)
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
			COALESCE(c.description, ''),
			c.created_by, 
			c.created_at, 
			c.updated_at,
			cm.chat_id,
			cm.user_id,
			cm.role,
			cm.joined_at
		FROM chats AS c
		JOIN chat_members AS user_members ON user_members.chat_id = c.id
		JOIN chat_members AS cm ON cm.chat_id = c.id
		WHERE user_members.user_id = $1
		ORDER BY c.updated_at DESC, c.id, cm.user_id`,
		userID,
	)
	if err != nil {
		log.Printf("[ERROR] repository get chats: %v", err)
		return nil, fmt.Errorf("get chats: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var c model.Chat
		var member model.ChatMember
		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Description,
			&c.CreatedBy,
			&c.CreatedAt,
			&c.UpdatedAt,
			&member.ChatID,
			&member.UserID,
			&member.Role,
			&member.JoinedAt,
		); err != nil {
			log.Printf("[ERROR] repository get chats: %v", err)
			return nil, fmt.Errorf("get chats: %w", err)
		}

		if len(chats) == 0 || chats[len(chats)-1].ID != c.ID {
			c.Members = make([]model.ChatMember, 0)
			chats = append(chats, c)
		}
		chats[len(chats)-1].Members = append(chats[len(chats)-1].Members, member)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] repository get chats: %v", err)
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
			COALESCE(description, ''),
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
			return nil, ErrChatNotFound
		}
		log.Printf("[ERROR] repository get chat info: %v", err)
		return nil, fmt.Errorf("get chat info: %w", err)
	}
	c.Members, err = getChatMembers(ctx, r.pool, c.ID)
	if err != nil {
		log.Printf("[ERROR] repository get chat info: %v", err)
		return nil, fmt.Errorf("get chat info: %w", err)
	}

	return &c, nil
}

func (r *pgChatsRepo) GetRoleInChat(
	ctx context.Context,
	chatID int64,
	userID int64,
) (*string, error) {
	var member model.ChatMember

	err := r.pool.QueryRow(ctx,
		`SELECT role
		FROM chat_members
		WHERE chat_id = $1 AND user_id = $2`,
		chatID, userID,
	).Scan(&member.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrChatMemberNotFound
		}
		log.Printf("[ERROR] repository get role in chat: %v", err)
		return nil, fmt.Errorf("get role in chat: %w", err)
	}

	return &member.Role, nil
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
			COALESCE(description, ''),
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
			return nil, ErrChatNotFound
		}
		log.Printf("[ERROR] repository update chat description: %v", err)
		return nil, fmt.Errorf("update chat: %w", err)
	}
	c.Members, err = getChatMembers(ctx, r.pool, c.ID)
	if err != nil {
		log.Printf("[ERROR] repository update chat description: %v", err)
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
			COALESCE(description, ''),
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
			return nil, ErrChatNotFound
		}
		log.Printf("[ERROR] repository update chat name: %v", err)
		return nil, fmt.Errorf("update chat: %w", err)
	}
	c.Members, err = getChatMembers(ctx, r.pool, c.ID)
	if err != nil {
		log.Printf("[ERROR] repository update chat name: %v", err)
		return nil, fmt.Errorf("update chat: %w", err)
	}
	return &c, nil
}

func getChatMembers(
	ctx context.Context,
	queryer chatMemberQueryer,
	chatID int64,
) ([]model.ChatMember, error) {
	members := make([]model.ChatMember, 0)
	rows, err := queryer.Query(ctx,
		`SELECT chat_id, user_id, role, joined_at
		FROM chat_members
		WHERE chat_id = $1
		ORDER BY user_id`,
		chatID,
	)
	if err != nil {
		return nil, fmt.Errorf("get chat members: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var member model.ChatMember
		if err := rows.Scan(
			&member.ChatID,
			&member.UserID,
			&member.Role,
			&member.JoinedAt,
		); err != nil {
			return nil, fmt.Errorf("get chat members: %w", err)
		}
		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get chat members: %w", err)
	}

	return members, nil
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
		log.Printf("[ERROR] repository delete chat: %v", err)
		return fmt.Errorf("delete chat: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrChatNotFound
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
		log.Printf("[ERROR] repository add member: %v", err)
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
		log.Printf("[ERROR] repository remove member: %v", err)
		return fmt.Errorf("remove member: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrChatMemberNotFound
	}

	return nil
}
