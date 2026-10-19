package service

import (
	"context"

	"github.com/ak-softwaredev/akrobat-msg/internal/model"
)

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

type chatsService struct {
	chats ChatsRepository
}

func NewChatsService(
	chats ChatsRepository,
) *chatsService {
	return &chatsService{
		chats: chats,
	}
}
