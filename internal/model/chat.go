package model

import "time"

type ChatRole string

const (
	RoleOwner  ChatRole = "owner"
	RoleAdmin  ChatRole = "admin"
	RoleMember ChatRole = "member"
)

type Chat struct {
	ID          int64        `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Members     []ChatMember `json:"members"`
	CreatedBy   *int64       `json:"created_by"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type ChatMember struct {
	ChatID   int64     `json:"chat_id"`
	UserID   int64     `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}
