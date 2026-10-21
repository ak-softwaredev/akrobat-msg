package handler

import (
	"context"
	"encoding/json"
	"encoding/json"
	"errors"
	"log"
	"math"
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

type MessagePublisher interface {
	Publish(chatID int64, data []byte)
}

type MessagePublisher interface {
	Publish(chatID int64, data []byte)
}

type messageHandler struct {
	messages MessagesService
	hub      MessagePublisher
	hub      MessagePublisher
}

func NewMessagesHandler(
	messages MessagesService,
	hub MessagePublisher,
	hub MessagePublisher,
) *messageHandler {
	return &messageHandler{
		messages: messages,
		hub:      hub,
		hub:      hub,
	}
}

type messageEvent struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func (h *messageHandler) publishEvent(
	chatID int64,
	eventType string,
	data any,
) {
	payload, err := json.Marshal(messageEvent{
		Type: eventType,
		Data: data,
	})
	if err != nil {
		log.Printf("[ERROR] marshal websocket event: %v", err)
		return
	}

	h.hub.Publish(chatID, payload)
type messageEvent struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func (h *messageHandler) publishEvent(
	chatID int64,
	eventType string,
	data any,
) {
	payload, err := json.Marshal(messageEvent{
		Type: eventType,
		Data: data,
	})
	if err != nil {
		log.Printf("[ERROR] marshal websocket event: %v", err)
		return
	}

	h.hub.Publish(chatID, payload)
}

type MessageJSON struct {
	Content string `json:"content"`
}

func (h *messageHandler) GetMessages(c *gin.Context) {
	userID := c.GetInt64("userID")
	chatID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	chatID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": gin.H{"message": "Invalid ChatID"}})
		return
	}

	beforeID, err := strconv.ParseInt(c.Query("beforeID"), 10, 64)
	if err != nil || beforeID <= 0 {
		beforeID = 0
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		limit = 50
		limit = 50
	}

	messages, err := h.messages.GetMessages(
		c.Request.Context(),
		chatID,
		userID,
		beforeID,
		limit,
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
	chatID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	chatID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": gin.H{"message": "Invalid ChatID"}})
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

	h.publishEvent(chatID, "message.created", message)

	h.publishEvent(chatID, "message.created", message)

	c.JSON(http.StatusCreated, message)
}

func (h *messageHandler) UpdateMessage(c *gin.Context) {
	userID := c.GetInt64("userID")
	chatID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	chatID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": gin.H{"message": "Invalid ChatID"}})
		return
	}
	messageID, err := strconv.ParseInt(c.Param("mid"), 10, 64)
	messageID, err := strconv.ParseInt(c.Param("mid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": gin.H{"message": "MessageID not a number"}})
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

	h.publishEvent(chatID, "message.updated", message)

	h.publishEvent(chatID, "message.updated", message)

	c.JSON(http.StatusOK, message)
}

func (h *messageHandler) DeleteMessage(c *gin.Context) {
	userID := c.GetInt64("userID")
	chatID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	chatID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": gin.H{"message": "Invalid ChatID"}})
		return
	}
	messageID, err := strconv.ParseInt(c.Param("mid"), 10, 64)
	messageID, err := strconv.ParseInt(c.Param("mid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": gin.H{"message": "MessageID not a number"}})
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

	h.publishEvent(chatID, "message.deleted", gin.H{
		"id":      messageID,
		"chat_id": chatID,
	})

	h.publishEvent(chatID, "message.deleted", gin.H{
		"id":      messageID,
		"chat_id": chatID,
	})

	c.Status(http.StatusNoContent)
}
