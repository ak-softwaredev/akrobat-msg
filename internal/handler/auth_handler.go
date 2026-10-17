package handler

import (
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

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

type AuthHandler interface {
	AuthRegister(c *gin.Context)
	AuthLogin(c *gin.Context)
	AuthLogout(c *gin.Context)
	// AuthRefresh(c *gin.Context)
}

type authHandler struct {
	auth service.AuthService
}

func NewAuthHandler(
	auth service.AuthService,
) AuthHandler {
	return &authHandler{
		auth: auth,
	}
}

func (a *authHandler) AuthRegister(c *gin.Context) {
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

	usernameLength := utf8.RuneCountInString(body.Username)
	if usernameLength < 3 || usernameLength > 20 {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": gin.H{"message": "Username must be 3-20 characters long"}},
		)
		return
	}
	passwordLength := utf8.RuneCountInString(body.Password)
	if passwordLength < 8 || passwordLength > 64 {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": gin.H{"message": "Password must be 8-64 characters long"}},
		)
		return
	}

	result, err := a.auth.Register(c.Request.Context(), body.Username, body.Password)
	if err != nil {
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

func (a *authHandler) AuthLogin(c *gin.Context) {
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

	uLen := utf8.RuneCountInString(body.Username)
	pLen := utf8.RuneCountInString(body.Password)
	if uLen < 3 || uLen > 20 || pLen < 8 || pLen > 64 {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": gin.H{"message": "Invalid request body"}},
		)
		return
	}

	result, err := a.auth.Login(c.Request.Context(), body.Username, body.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{"error": gin.H{"message": "Invalid credentials"}},
			)
			return
		}
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

func (a *authHandler) AuthLogout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")

	if err := a.auth.Logout(c.Request.Context(), token); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"errors": gin.H{"message": "Invalid token"}},
		)
		return
	}

	c.JSON(
		http.StatusAccepted,
		gin.H{"message": "Logged out successfully", "status": "ok"},
	)
}
