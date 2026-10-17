package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ak-softwaredev/akrobat-msg/internal/service"
	"github.com/gin-gonic/gin"
)

const UserIDKey = "userID"

func Auth(s service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		token = strings.TrimPrefix(token, "Bearer ")

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
