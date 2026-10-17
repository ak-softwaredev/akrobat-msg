package main

import (
	"context"
	"log"
	"net/http"
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

	UsersRepository := repository.NewUsersRepository(pool)
	SessionRepository := repository.NewSessionRepository(pool)
	AuthService := service.NewAuthService(UsersRepository, SessionRepository)
	AuthHandler := handler.NewAuthHandler(AuthService)

	r.POST("/api/v1/auth/register", AuthHandler.AuthRegister)
	r.POST("/api/v1/auth/login", AuthHandler.AuthLogin)
	r.POST("/api/v1/auth/logout", AuthHandler.AuthLogout)

	api := r.Group("/api")
	api.Use(middleware.Auth(AuthService))

	r.GET("/health/live", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

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
