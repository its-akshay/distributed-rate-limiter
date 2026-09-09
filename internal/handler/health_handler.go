package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewHealthHandler(
	db *pgxpool.Pool,
	redis *redis.Client,
) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

// Health godoc
// @Summary Liveness check
// @Description Returns ok once the process is up
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": "v5",
	})
}

// Ready godoc
// @Summary Readiness check
// @Description Checks Postgres and Redis connectivity
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {

	err := h.db.Ping(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "postgres unavailable",
		})
		return
	}

	err = h.redis.Ping(c.Request.Context()).Err()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "redis unavailable",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}
