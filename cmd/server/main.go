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
	"github.com/ak-softwaredev/akrobat-msg/internal/ws"
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

	log.Println("starting migrations")
	if err := database.RunMigrations(os.Getenv("DATABASE_URL")); err != nil {
		os.Exit(1)
	}
	log.Println("migrations complete")

	log.Println("creating database pool")
	pool, err := database.NewPool(ctx)
	if err != nil {
		os.Exit(1)
	}
	defer pool.Close()
	log.Println("database pool ready")

	log.Println("registering endpoints")
	usersRepository := repository.NewUsersRepository(pool)
	sessionRepository := repository.NewSessionRepository(pool)
	chatsRepository := repository.NewChatsRepository(pool)
	messagesRepository := repository.NewMessagesRepository(pool)

	usersService := service.NewUsersService(usersRepository)
	authService := service.NewAuthService(usersRepository, sessionRepository)
	chatsService := service.NewChatsService(chatsRepository)
	messagesService := service.NewMessageService(chatsRepository, messagesRepository)

	hub := ws.NewHub()
	go hub.Run()

	usersHandler := handler.NewUsersHandler(usersService)
	authHandler := handler.NewAuthHandler(authService)
	chatsHandler := handler.NewChatHandler(chatsService)
	messagesHandler := handler.NewMessagesHandler(messagesService, hub)
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

	chatsv1 := r.Group("/api/v1/chats")
	chatsv1.Use(
		middleware.Auth(authService),
		middleware.CheckUserInChat(chatsService),
	)

	// chats
	api.GET("/v1/chats", chatsHandler.GetChats)
	api.POST("/v1/chats", chatsHandler.CreateChat)
	chatsv1.GET("/:id", chatsHandler.GetChatInfo)
	chatsv1.PATCH("/:id", chatsHandler.PatchChat)
	chatsv1.DELETE("/:id", chatsHandler.DeleteChat)
	chatsv1.POST("/:id/members", chatsHandler.AddMember)
	chatsv1.DELETE("/:id/members/:rid", chatsHandler.RemoveMember)

	// messages
	chatsv1.GET("/:id/messages", messagesHandler.GetMessages)
	chatsv1.POST("/:id/messages", messagesHandler.PostMessage)
	chatsv1.PATCH("/:id/messages/:mid", messagesHandler.UpdateMessage)
	chatsv1.DELETE("/:id/messages/:mid", messagesHandler.DeleteMessage)

	// ws upgrade
	chatsv1.GET("/:id/ws", hub.HandleChat)
	log.Println("registering complete")

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}
	log.Printf("port: %s\n", port)

	if err := r.Run(port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
