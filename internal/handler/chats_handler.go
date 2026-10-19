package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/ak-softwaredev/akrobat-msg/internal/model"
	"github.com/ak-softwaredev/akrobat-msg/internal/service"
	"github.com/gin-gonic/gin"
)

type ChatsService interface {
	CreateChat(ctx context.Context, userID int64, name string) (*model.Chat, error)
	GetChats(ctx context.Context, userID int64) ([]model.Chat, error)
	GetChatDetails(ctx context.Context, chatID, userID int64) (*model.Chat, error)
	UpdateChat(ctx context.Context, chatID, userID int64, field, value string) (*model.Chat, error)
	DeleteChat(ctx context.Context, chatID, userID int64) error
	AddMember(ctx context.Context, chatID, userID int64, addID int64) error
	RemoveMember(ctx context.Context, chatID, userID, removedID int64) error
}

type CreateChatJSON struct {
	Name string `json:"name"`
}

type UpdateChatJSON struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

type AddMemberJSON struct {
	UserID int64 `json:"id"`
}

type chatHandler struct {
	chats ChatsService
}

func NewChatHandler(chats ChatsService) *chatHandler {
	return &chatHandler{
		chats: chats,
	}
}

func (h *chatHandler) CreateChat(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		64<<10, // 64 * 2^10, 64 KiB
	)

	userID := c.GetInt64("userID")

	body := CreateChatJSON{}
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Printf("[ERROR] handler bind create chat request: %v", err)
		c.JSON(http.StatusBadRequest,
			gin.H{"error": gin.H{"message": "invalid json"}})
		return
	}

	chat, err := h.chats.CreateChat(c.Request.Context(), userID, body.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}

	c.JSON(http.StatusCreated, chat)
}

func (h *chatHandler) GetChats(c *gin.Context) {
	userID := c.GetInt64("userID")

	chats, err := h.chats.GetChats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}

	c.JSON(http.StatusOK, chats)
}

func (h *chatHandler) GetChatInfo(c *gin.Context) {
	userID := c.GetInt64("userID")
	stringChatID := c.Param("id")
	chatID, err := strconv.ParseInt(stringChatID, 10, 0)
	if err != nil {
		log.Printf("[ERROR] handler parse chat id: %v", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}

	chat, err := h.chats.GetChatDetails(c.Request.Context(), chatID, userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized,
				gin.H{"error": gin.H{"message": "Insufficient Permissions"}})
			return
		}
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}

	c.JSON(http.StatusOK, chat)
}

func (h *chatHandler) PatchChat(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		64<<10, // 64 * 2^10, 64 KiB
	)

	userID := c.GetInt64("userID")
	stringChatID := c.Param("id")
	chatID, err := strconv.ParseInt(stringChatID, 10, 0)
	if err != nil {
		log.Printf("[ERROR] handler parse chat id: %v", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}

	body := UpdateChatJSON{}
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Printf("[ERROR] handler bind update chat request: %v", err)
		c.JSON(http.StatusBadRequest,
			gin.H{"error": gin.H{"message": "invalid json"}})
		return
	}

	chat, err := h.chats.UpdateChat(
		c.Request.Context(),
		chatID, userID,
		body.Field, body.Value,
	)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized,
				gin.H{"error": gin.H{"message": "Insufficient Permissions"}})
			return
		}
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}

	c.JSON(http.StatusOK, chat)
}

func (h *chatHandler) DeleteChat(c *gin.Context) {
	userID := c.GetInt64("userID")
	stringChatID := c.Param("id")
	chatID, err := strconv.ParseInt(stringChatID, 10, 0)
	if err != nil {
		log.Printf("[ERROR] handler parse chat id: %v", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}

	err = h.chats.DeleteChat(c.Request.Context(), chatID, userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized,
				gin.H{"error": gin.H{"message": "Insufficient Permissions"}})
			return
		}
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *chatHandler) AddMember(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		64<<10, // 64 * 2^10, 64 KiB
	)

	userID := c.GetInt64("userID")
	stringChatID := c.Param("id")
	chatID, err := strconv.ParseInt(stringChatID, 10, 0)
	if err != nil {
		log.Printf("[ERROR] handler parse chat id: %v", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}

	body := AddMemberJSON{}
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Printf("[ERROR] handler bind update chat request: %v", err)
		c.JSON(http.StatusBadRequest,
			gin.H{"error": gin.H{"message": "invalid json"}})
		return
	}

	err = h.chats.AddMember(c.Request.Context(), chatID, userID, body.UserID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized,
				gin.H{"error": gin.H{"message": "Insufficient Permissions"}})
			return
		}
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *chatHandler) RemoveMember(c *gin.Context) {
	userID := c.GetInt64("userID")
	stringChatID := c.Param("id")
	chatID, err := strconv.ParseInt(stringChatID, 10, 0)
	if err != nil {
		log.Printf("[ERROR] handler parse chat id: %v", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}
	stringRemovedID := c.Param("rid")
	removedID, err := strconv.ParseInt(stringRemovedID, 10, 0)
	if err != nil {
		log.Printf("[ERROR] handler parse removed user id: %v", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}

	err = h.chats.RemoveMember(c.Request.Context(), chatID, userID, removedID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized,
				gin.H{"error": gin.H{"message": "Insufficient Permissions"}})
			return
		}
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}

	c.Status(http.StatusNoContent)
}

/*
[-] POST /api/v1/chats <- name
[-] GET /api/v1/chats -> all chats
[-] GET /api/v1/chats/:id -> chat details
[-] PATCH /api/v1/chats/:id <- name, description
[-] DELETE /api/v1/chats/:id -> nil
[-] POST /api/v1/chats/:id/members <- userid
[-] DELETE /api/v1/chats/:id/members/:id -> nil
*/
