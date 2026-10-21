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

type MessagesService interface {
	GetMessages(ctx context.Context, chatID, userID int64, beforeID int64, limit int) ([]model.Message, error)
	PostMessage(ctx context.Context, chatID, userID int64, content string) (*model.Message, error)
	UpdateMessage(ctx context.Context, chatID, userID, messageID int64, content string) (*model.Message, error)
	DeleteMessage(ctx context.Context, chatID, userID, messageID int64) error
}

type messageHandler struct {
	messages MessagesService
}

func NewMessagesHandler(
	messages MessagesService,
) *messageHandler {
	return &messageHandler{
		messages: messages,
	}
}

type GetMessagesJSON struct {
	BeforeID int64 `json:"before_id"`
	Limit    int   `json:"limit"`
}

type MessageJSON struct {
	Content string `json:"content"`
}

func (h *messageHandler) GetMessages(c *gin.Context) {
	userID := c.GetInt64("userID")
	stringChatID := c.Param("id")
	chatID, err := strconv.ParseInt(stringChatID, 10, 0)
	if err != nil {
		log.Printf("[ERROR] handler parse chat id: %v", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}

	body := GetMessagesJSON{}
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Printf("[ERROR] handler bind get messages request: %v", err)
		c.JSON(http.StatusBadRequest,
			gin.H{"error": gin.H{"message": "invalid json"}})
		return
	}

	messages, err := h.messages.GetMessages(
		c.Request.Context(),
		chatID,
		userID,
		body.BeforeID,
		body.Limit,
	)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) ||
			errors.Is(err, service.ErrInsufficientPermissions) {
			c.JSON(http.StatusUnauthorized,
				gin.H{"error": gin.H{"message": "Insufficient Permissions"}})
			return
		}
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (h *messageHandler) PostMessage(c *gin.Context) {
	userID := c.GetInt64("userID")
	stringChatID := c.Param("id")
	chatID, err := strconv.ParseInt(stringChatID, 10, 0)
	if err != nil {
		log.Printf("[ERROR] handler parse chat id: %v", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}

	body := MessageJSON{}
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Printf("[ERROR] handler bind post message request: %v", err)
		c.JSON(http.StatusBadRequest,
			gin.H{"error": gin.H{"message": "invalid json"}})
		return
	}

	message, err := h.messages.PostMessage(
		c.Request.Context(),
		chatID,
		userID,
		body.Content,
	)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) ||
			errors.Is(err, service.ErrInsufficientPermissions) {
			c.JSON(http.StatusUnauthorized,
				gin.H{"error": gin.H{"message": "Insufficient Permissions"}})
			return
		}
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}

	c.JSON(http.StatusCreated, message)
}

func (h *messageHandler) UpdateMessage(c *gin.Context) {
	userID := c.GetInt64("userID")
	stringChatID := c.Param("id")
	chatID, err := strconv.ParseInt(stringChatID, 10, 0)
	if err != nil {
		log.Printf("[ERROR] handler parse chat id: %v", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}
	stringMessageID := c.Param("mid")
	messageID, err := strconv.ParseInt(stringMessageID, 10, 0)
	if err != nil {
		log.Printf("[ERROR] handler parse message id: %v", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}

	body := MessageJSON{}
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Printf("[ERROR] handler bind update message request: %v", err)
		c.JSON(http.StatusBadRequest,
			gin.H{"error": gin.H{"message": "invalid json"}})
		return
	}

	message, err := h.messages.UpdateMessage(
		c.Request.Context(),
		chatID,
		userID,
		messageID,
		body.Content,
	)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) ||
			errors.Is(err, service.ErrInsufficientPermissions) {
			c.JSON(http.StatusUnauthorized,
				gin.H{"error": gin.H{"message": "Insufficient Permissions"}})
			return
		}
		if errors.Is(err, service.ErrMessageNotFound) {
			c.JSON(http.StatusNotFound,
				gin.H{"error": gin.H{"message": "Message not found"}})
			return
		}
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}

	c.JSON(http.StatusOK, message)
}

func (h *messageHandler) DeleteMessage(c *gin.Context) {
	userID := c.GetInt64("userID")
	stringChatID := c.Param("id")
	chatID, err := strconv.ParseInt(stringChatID, 10, 0)
	if err != nil {
		log.Printf("[ERROR] handler parse chat id: %v", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}
	stringMessageID := c.Param("mid")
	messageID, err := strconv.ParseInt(stringMessageID, 10, 0)
	if err != nil {
		log.Printf("[ERROR] handler parse message id: %v", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}

	err = h.messages.DeleteMessage(
		c.Request.Context(),
		chatID,
		userID,
		messageID,
	)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) ||
			errors.Is(err, service.ErrInsufficientPermissions) {
			c.JSON(http.StatusUnauthorized,
				gin.H{"error": gin.H{"message": "Insufficient Permissions"}})
			return
		}
		if errors.Is(err, service.ErrMessageNotFound) {
			c.JSON(http.StatusNotFound,
				gin.H{"error": gin.H{"message": "Message not found"}})
			return
		}
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}})
		return
	}

	c.Status(http.StatusNoContent)
}
