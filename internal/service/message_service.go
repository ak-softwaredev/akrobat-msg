package service

import (
	"context"

	"github.com/ak-softwaredev/akrobat-msg/internal/model"
)

type MessagesRepository interface {
	GetChatMessages(ctx context.Context, chatID int64, beforeID int64, limit int) ([]model.Message, error)
	CreateMessage(ctx context.Context, chatID int64, userID int64, content string) (*model.Message, error)
	UpdateMessage(ctx context.Context, chatID int64, messageID int64, userID int64, content string) (*model.Message, error)
	DeleteMessage(ctx context.Context, chatID int64, messageID int64, userID int64) error
}
