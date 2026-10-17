package model

import "time"

type Message struct {
	ID        int64     `json:"id"`
	ChatID    int64     `json:"chat_id"`
	SenderID  *int64    `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
