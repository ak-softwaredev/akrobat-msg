package service

import (
	"context"
	"errors"

	"github.com/ak-softwaredev/akrobat-msg/internal/model"
	"github.com/ak-softwaredev/akrobat-msg/internal/repository"
)

var (
	ErrInvalidRoomLength       = errors.New("Room name must be 1-255 characters long")
	ErrChatNotFound            = errors.New("chat not found")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrInvalidChatFiend        = errors.New("invalid chat field")
)

type ChatsRepository interface {
	CreateChat(ctx context.Context, userID int64, name string) (*model.Chat, error)
	GetUserChats(ctx context.Context, userID int64) ([]model.Chat, error)
	GetRoleInChat(ctx context.Context, chatID int64, userID int64) (*string, error)
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

func (s *chatsService) CreateChat(
	ctx context.Context,
	userID int64,
	name string,
) (*model.Chat, error) {
	nameLength := len(name)
	if nameLength < 1 || nameLength > 255 {
		return nil, ErrInvalidRoomLength
	}

	c, err := s.chats.CreateChat(ctx, userID, name)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s *chatsService) GetChats(
	ctx context.Context,
	userID int64,
) ([]model.Chat, error) {
	chats, err := s.chats.GetUserChats(ctx, userID)
	if err != nil {
		return nil, err
	}

	return chats, err
}

func (s *chatsService) GetChatDetails(
	ctx context.Context,
	chatID, userID int64,
) (*model.Chat, error) {
	_, err := s.chats.GetRoleInChat(ctx, chatID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrChatMemberNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	chat, err := s.chats.GetChatInfo(ctx, chatID)
	if err != nil {
		if errors.Is(err, repository.ErrChatNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return chat, err
}

func (s *chatsService) UpdateChat(
	ctx context.Context,
	chatID, userID int64,
	field, value string,
) (*model.Chat, error) {
	var c *model.Chat
	var err error

	rank, err := s.chats.GetRoleInChat(ctx, chatID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrChatMemberNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if *rank != "admin" && *rank != "owner" {
		return nil, ErrInsufficientPermissions
	}

	switch field {
	case "name":
		nameLength := len(value)
		if nameLength < 1 || nameLength > 255 {
			return nil, ErrInvalidRoomLength
		}

		c, err = s.chats.UpdateChatName(ctx, chatID, value)
	case "description":
		c, err = s.chats.UpdateChatDescription(ctx, chatID, value)
	default:
		err = ErrInvalidChatFiend
	}

	if err != nil {
		if errors.Is(err, repository.ErrChatNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return c, nil
}

func (s *chatsService) DeleteChat(
	ctx context.Context,
	chatID, userID int64,
) error {
	rank, err := s.chats.GetRoleInChat(ctx, chatID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrChatMemberNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if *rank != "owner" {
		return ErrInsufficientPermissions
	}

	err = s.chats.DeleteChat(ctx, chatID)
	if err != nil {
		return err
	}

	return nil
}

func (s *chatsService) AddMember(
	ctx context.Context,
	chatID, userID, addID int64,
) error {
	rank, err := s.chats.GetRoleInChat(ctx, chatID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrChatMemberNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if *rank != "admin" && *rank != "owner" {
		return ErrInsufficientPermissions
	}

	err = s.chats.AddMember(ctx, chatID, addID)
	if err != nil {
		// double-check, idk why, better be safe than sorry!
		if errors.Is(err, repository.ErrChatNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}

func (s *chatsService) RemoveMember(
	ctx context.Context,
	chatID, userID, removedID int64,
) error {
	rank, err := s.chats.GetRoleInChat(ctx, chatID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrChatMemberNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if *rank != "admin" && *rank != "owner" {
		return ErrInsufficientPermissions
	}

	err = s.chats.RemoveMember(ctx, chatID, removedID)
	if err != nil {
		// same here! still dont know why!
		if errors.Is(err, repository.ErrChatNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}
