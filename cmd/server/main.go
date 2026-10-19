package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/ak-softwaredev/akrobat-msg/internal/database"
	"github.com/ak-softwaredev/akrobat-msg/internal/handler"
	"github.com/ak-softwaredev/akrobat-msg/internal/middleware"
	"github.com/ak-softwaredev/akrobat-msg/internal/repository"
	"github.com/ak-softwaredev/akrobat-msg/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("[ERROR] load environment: %v", err)
	}
	ctx := context.Background()
	r := gin.Default()
	r.SetTrustedProxies(nil)

	if err := database.RunMigrations(os.Getenv("DATABASE_URL")); err != nil {
		os.Exit(1)
	}

	pool, err := database.NewPool(ctx)
	if err != nil {
		os.Exit(1)
	}
	defer pool.Close()

	usersRepository := repository.NewUsersRepository(pool)
	sessionRepository := repository.NewSessionRepository(pool)
	chatsRepository := repository.NewChatsRepository(pool)

	usersService := service.NewUsersService(usersRepository)
	authService := service.NewAuthService(usersRepository, sessionRepository)
	chatsService := service.NewChatsService(chatsRepository)

	usersHandler := handler.NewUsersHandler(usersService)
	authHandler := handler.NewAuthHandler(authService)
	chatsHandler := handler.NewChatHandler(chatsService)
	healthHandler := handler.NewHealthHandler(pool)

	// health
	r.GET("/health/live", healthHandler.HealthLive)
	r.GET("/health/ready", healthHandler.HealthReady)

	// auth
	r.POST("/api/v1/auth/register", authHandler.AuthRegister)
	r.POST("/api/v1/auth/login", authHandler.AuthLogin)
	r.POST("/api/v1/auth/logout", authHandler.AuthLogout)

	api := r.Group("/api")
	api.Use(middleware.Auth(authService))

	// users
	api.GET("/v1/users/me", usersHandler.GetSelf)
	api.GET("/v1/users/:id", usersHandler.GetByID)
	api.PATCH("/v1/users/me", usersHandler.Update)
	api.DELETE("/v1/users/me", usersHandler.Delete)

	// chats
	api.GET("/v1/chats", chatsHandler.GetChats)
	api.POST("/v1/chats", chatsHandler.CreateChat)
	api.GET("/v1/chats/:id", chatsHandler.GetChatInfo)
	api.PATCH("/v1/chats/:id", chatsHandler.PatchChat)
	api.DELETE("/v1/chats/:id", chatsHandler.DeleteChat)
	api.POST("/v1/chats/:id/members", chatsHandler.AddMember)
	api.DELETE("/v1/chats/:id/members/:rid", chatsHandler.RemoveMember)

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}

	if err := r.RunTLS(
		port,
		"certs/localhost.pem",
		"certs/localhost-key.pem",
	); err != nil {
		log.Fatalf("server: %v", err)
	}
}
