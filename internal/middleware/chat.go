package middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ChatsService interface {
	IsUserInChat(ctx context.Context, chatID, userID int64) (bool, error)
}

func CheckUserInChat(s ChatsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("userID")
		chatID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": gin.H{"message": "Invalid ChatID"}})
			return
		}

		isInChat, err := s.IsUserInChat(c.Request.Context(), chatID, userID)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"error": "Internal server error"},
			)
			return
		}
		if !isInChat {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{"error": "Cannot access chat"},
			)
			return
		}

		c.Next()
	}
}
