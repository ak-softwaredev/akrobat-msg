package handler

import (
	"context"

	"github.com/ak-softwaredev/akrobat-msg/internal/model"
)

type MessagesService interface {
	GetMessages(ctx context.Context, chatID, userID int64, beforeID int64, limit int) ([]model.Message, error)
	PostMessage(ctx context.Context, chatID, userID int64, content string) (*model.Message, error)
	UpdateMessage(ctx context.Context, chatID, userID, messageID int64, content string) (*model.Message, error)
	DeleteMessage(ctx context.Context, chatID, userID, messageID int64) error
}
