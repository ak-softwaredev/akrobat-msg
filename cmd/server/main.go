package main

import (
	"context"
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
	godotenv.Load()
	ctx := context.Background()
	r := gin.Default()
	r.SetTrustedProxies(nil)

	if err := database.RunMigrations(os.Getenv("DATABASE_URL")); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	usersRepository := repository.NewUsersRepository(pool)
	sessionRepository := repository.NewSessionRepository(pool)
	authService := service.NewAuthService(usersRepository, sessionRepository)
	authHandler := handler.NewAuthHandler(authService)
	healthHandler := handler.NewHealthHandler(pool)

	r.GET("/health/live", healthHandler.HealthLive)
	r.GET("/health/ready", healthHandler.HealthReady)

	r.POST("/api/v1/auth/register", authHandler.AuthRegister)
	r.POST("/api/v1/auth/login", authHandler.AuthLogin)
	r.POST("/api/v1/auth/logout", authHandler.AuthLogout)

	api := r.Group("/api")
	api.Use(middleware.Auth(authService))

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
