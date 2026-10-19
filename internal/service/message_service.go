package service

import (
	"context"
	"errors"

	"github.com/ak-softwaredev/akrobat-msg/internal/model"
	"github.com/ak-softwaredev/akrobat-msg/internal/repository"
)

var ErrMessageNotFound = errors.New("message not found")

type MessagesRepository interface {
	GetChatMessages(ctx context.Context, chatID int64, beforeID int64, limit int) ([]model.Message, error)
	CreateMessage(ctx context.Context, chatID int64, userID int64, content string) (*model.Message, error)
	UpdateMessage(ctx context.Context, chatID int64, userID int64, messageID int64, content string) (*model.Message, error)
	DeleteMessage(ctx context.Context, chatID int64, userID int64, messageID int64) error
}

type messageService struct {
	chats    ChatsRepository
	messages MessagesRepository
}

func NewMessageService(
	chats ChatsRepository,
	messages MessagesRepository,
) *messageService {
	return &messageService{
		chats:    chats,
		messages: messages,
	}
}

func (s *messageService) GetMessages(
	ctx context.Context,
	chatID, userID int64,
	beforeID int64,
	limit int,
) ([]model.Message, error) {
	_, err := s.chats.GetRoleInChat(ctx, chatID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrChatMemberNotFound) {
			return nil, ErrInsufficientPermissions
		}
		return nil, err
	}

	msgs, err := s.messages.GetChatMessages(ctx, chatID, beforeID, limit)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

func (s *messageService) PostMessage(
	ctx context.Context,
	chatID, userID int64,
	content string,
) (*model.Message, error) {
	_, err := s.chats.GetRoleInChat(ctx, chatID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrChatMemberNotFound) {
			return nil, ErrInsufficientPermissions
		}
		return nil, err
	}

	msg, err := s.messages.CreateMessage(ctx, chatID, userID, content)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *messageService) UpdateMessage(
	ctx context.Context,
	chatID, userID, messageID int64,
	content string,
) (*model.Message, error) {
	_, err := s.chats.GetRoleInChat(ctx, chatID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrChatMemberNotFound) {
			return nil, ErrInsufficientPermissions
		}
		return nil, err
	}

	msg, err := s.messages.UpdateMessage(ctx, chatID, userID, messageID, content)
	if err != nil {
		if errors.Is(err, repository.ErrMessageNotFound) {
			return nil, ErrMessageNotFound
		}
		return nil, err
	}

	return msg, nil
}

func (s *messageService) DeleteMessage(
	ctx context.Context,
	chatID, userID, messageID int64,
) error {
	_, err := s.chats.GetRoleInChat(ctx, chatID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrChatMemberNotFound) {
			return ErrInsufficientPermissions
		}
		return err
	}

	err = s.messages.DeleteMessage(ctx, chatID, userID, messageID)
	if err != nil {
		if errors.Is(err, repository.ErrMessageNotFound) {
			return ErrMessageNotFound
		}
		return err
	}

	return nil
}
