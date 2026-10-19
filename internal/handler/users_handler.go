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

type UsersService interface {
	GetByID(ctx context.Context, id int64) (*model.User, error)
	Update(ctx context.Context, id int64, field string, value string) (*model.User, error)
	Delete(ctx context.Context, id int64) error
}

type UpdateField struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

type usersHandler struct {
	users UsersService
}

func NewUsersHandler(
	users UsersService,
) *usersHandler {
	return &usersHandler{
		users: users,
	}
}

func (h *usersHandler) GetSelf(c *gin.Context) {
	id := c.GetInt64("userID")

	u, err := h.users.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": gin.H{"message": "user doesnt exist"}},
			)
			return
		}
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}},
		)
		return
	}

	c.JSON(http.StatusOK, u)
}

func (h *usersHandler) GetByID(c *gin.Context) {
	stringID := c.Param("id")
	id, err := strconv.ParseInt(stringID, 10, 0)
	if err != nil || id < 0 {
		if err != nil {
			log.Printf("[ERROR] handler parse user id: %v", err)
		}
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": gin.H{"message": "invalid id"}},
		)
		return
	}

	u, err := h.users.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": gin.H{"message": "user doesnt exist"}},
			)
			return
		}
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}},
		)
		return
	}

	c.JSON(http.StatusOK, u)
}

func (h *usersHandler) Update(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		64<<10, // 64 * 2^10, 64 KiB
	)

	id := c.GetInt64("userID")

	body := UpdateField{}
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Printf("[ERROR] handler bind update user request: %v", err)
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": gin.H{"message": "Invalid request body"}},
		)
		return
	}

	u, err := h.users.Update(
		c.Request.Context(),
		id,
		body.Field,
		body.Value,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUserField):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"message":   "Invalid user field",
					"available": []string{"username, password, bio"},
				},
			})
		case errors.Is(err, service.ErrUsernameBadLength):
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": gin.H{"message": "Username must be 3-20 characters"},
			})
		case errors.Is(err, service.ErrPasswordBadLength):
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": gin.H{"message": "Password must be 8-64 characters"},
			})
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": "User not found"},
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"message": "Internal Server Error"},
			})
		}
		return
	}

	c.JSON(http.StatusOK, u)
}

func (h *usersHandler) Delete(c *gin.Context) {
	id := c.GetInt64("userID")

	err := h.users.Delete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": gin.H{"message": "user doesnt exist"}},
			)
			return
		}
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}},
		)
		return
	}

	c.Status(http.StatusNoContent)
}
