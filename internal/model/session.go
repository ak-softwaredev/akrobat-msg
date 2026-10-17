package model

import "time"

type Session struct {
	TokenHash string    `json:"-"`
	UserID    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}
