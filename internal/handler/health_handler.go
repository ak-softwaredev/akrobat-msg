package handler

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler interface {
	HealthLive(c *gin.Context)
	HealthReady(c *gin.Context)
}

type healthHandler struct {
	pool *pgxpool.Pool
}

func NewHealthHandler(
	pool *pgxpool.Pool,
) HealthHandler {
	return &healthHandler{
		pool: pool,
	}
}

func (h *healthHandler) HealthLive(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		gin.H{"status": "ok"},
	)
}

func (h *healthHandler) HealthReady(c *gin.Context) {
	ctx, cancel := context.WithTimeout(
		c.Request.Context(),
		5*time.Second,
	)
	defer cancel()

	if err := h.pool.Ping(ctx); err != nil {
		log.Printf("[ERROR] health ready: %v", err)
		c.JSON(
			http.StatusServiceUnavailable,
			gin.H{"status": "down", "reason": "database unavailable"},
		)
		return
	}
	c.JSON(
		http.StatusOK,
		gin.H{"status": "ok"},
	)
}
