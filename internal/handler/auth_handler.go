package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/ak-softwaredev/akrobat-msg/internal/service"
	"github.com/gin-gonic/gin"
)

type InputJSON struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"auth_token"`
}

type OutputJSON struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Token    string `json:"auth_token"`
}

type AuthService interface {
	Register(ctx context.Context, username, password string) (*service.AuthSuccessResult, error)
	Login(ctx context.Context, username, password string) (*service.AuthSuccessResult, error)
	Logout(ctx context.Context, token string) error
}

type authHandler struct {
	auth AuthService
}

func NewAuthHandler(
	auth AuthService,
) *authHandler {
	return &authHandler{
		auth: auth,
	}
}

func (h *authHandler) AuthRegister(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		64<<10, // 64 * 2^10, 64 KiB
	)

	body := InputJSON{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": gin.H{"message": "Invalid request body"}},
		)
		return
	}

	result, err := h.auth.Register(c.Request.Context(), body.Username, body.Password)
	if err != nil {
		if errors.Is(err, service.ErrUsernameBadLength) {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": gin.H{"message": "Username must be 3-20 characters long"}},
			)
			return
		}
		if errors.Is(err, service.ErrPasswordBadLength) {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": gin.H{"message": "Password must be 8-64 characters long"}},
			)
			return
		}
		if errors.Is(err, service.ErrUsernameTaken) {
			c.JSON(
				http.StatusConflict,
				gin.H{"error": gin.H{"message": body.Username + " is taken"}},
			)
			return
		}
		log.Printf("[ERROR] auth register: %v", err)
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}},
		)
		return
	}

	response := OutputJSON{}
	response.ID = result.Session.UserID
	response.Username = body.Username
	response.Token = result.Token

	c.JSON(
		http.StatusCreated,
		response,
	)
}

func (h *authHandler) AuthLogin(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		64<<10, // 64 * 2^10, 64 KiB
	)

	body := InputJSON{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": gin.H{"message": "Invalid request body"}},
		)
		return
	}

	result, err := h.auth.Login(c.Request.Context(), body.Username, body.Password)
	if err != nil {
		if errors.Is(err, service.ErrUsernameBadLength) || errors.Is(err, service.ErrPasswordBadLength) {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": gin.H{"message": "Invalid request body"}},
			)
			return
		}
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{"error": gin.H{"message": "Invalid credentials"}},
			)
			return
		}
		log.Printf("[ERROR] auth login: %v", err)
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": gin.H{"message": "Internal Server Error"}},
		)
		return
	}

	response := OutputJSON{}
	response.ID = result.Session.UserID
	response.Username = body.Username
	response.Token = result.Token

	c.JSON(
		http.StatusOK,
		response,
	)
}

func (h *authHandler) AuthLogout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")

	if err := h.auth.Logout(c.Request.Context(), token); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"errors": gin.H{"message": "Invalid token"}},
		)
		return
	}

	c.Status(http.StatusNoContent)
}
