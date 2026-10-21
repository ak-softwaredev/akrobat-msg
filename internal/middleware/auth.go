package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/ak-softwaredev/akrobat-msg/internal/service"
	"github.com/gin-gonic/gin"
)

const UserIDKey = "userID"

type AuthService interface {
	Authenticate(ctx context.Context, token string) (int64, error)
}

func Auth(s AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")
		if token == "" && strings.HasSuffix(c.Request.URL.Path, "/ws") {
			token = c.Query("token")
		}

		userID, err := s.Authenticate(
			c.Request.Context(),
			token,
		)
		if err != nil {
			if errors.Is(err, service.ErrInvalidCredentials) {
				c.AbortWithStatusJSON(
					http.StatusUnauthorized,
					gin.H{"error": "Unauthorized"},
				)
				return
			}

			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"error": "Internal server error"},
			)
			return
		}

		c.Set(UserIDKey, userID)
		c.Next()
	}
}
