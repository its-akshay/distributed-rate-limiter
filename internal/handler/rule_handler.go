package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/its-akshay/distributed-rate-limiter/internal/metrics"
	"github.com/its-akshay/distributed-rate-limiter/internal/model"
	"github.com/its-akshay/distributed-rate-limiter/internal/repository"
	"github.com/its-akshay/distributed-rate-limiter/internal/service"
)

type RuleHandler struct {
	repo    *repository.RuleRepository
	service *service.RateLimiterService
}

func NewRuleHandler(repo *repository.RuleRepository, service *service.RateLimiterService) *RuleHandler {
	return &RuleHandler{
		repo:    repo,
		service: service,
	}
}

// CreateRule godoc
// @Summary Create rate limit rule
// @Description Create a new rate limiting rule
// @Tags Rules
// @Accept json
// @Produce json
// @Param rule body model.Rule true "Rule payload"
// @Success 201 {object} model.Rule
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /rules [post]
func (h *RuleHandler) CreateRule(c *gin.Context) {
	var rule model.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err := h.repo.Create(context.Background(), &rule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, rule)
}


// GetRule godoc
// @Summary Get rule by ID
// @Description Fetch a rule by ID
// @Tags Rules
// @Produce json
// @Param id path int true "Rule ID"
// @Success 200 {object} model.Rule
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Router /rules/{id} [get]
func (h *RuleHandler) GetRule(c *gin.Context) {
	idstr := c.Param("id")

	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id",
		})
		return
	}
	rule, err := h.repo.GetById(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "rule not found",
		})
		return
	}
	c.JSON(http.StatusOK, rule)
}

// ListRules godoc
// @Summary List all rules
// @Description Returns all configured rate limit rules
// @Tags Rules
// @Produce json
// @Success 200 {array} model.Rule
// @Failure 500 {object} model.ErrorResponse
// @Router /rules [get]
func (h *RuleHandler) ListRules(c *gin.Context) {
	rules, err := h.repo.List(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, rules)
}


// Check godoc
// @Summary Check rate limit
// @Description Evaluates whether a request is allowed
// @Tags Rate Limiter
// @Accept json
// @Produce json
// @Param request body model.CheckRequest true "Rate limit check request"
// @Success 200 {object} model.CheckResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /check [post]
func (h *RuleHandler) Check(c *gin.Context) {
	metrics.RequestsTotal.Inc()
	var req model.CheckRequest

	if err:=c.ShouldBindJSON(&req); err!=nil {
		metrics.ErrorsTotal.Inc()
		c.JSON(http.StatusBadRequest, gin.H{
			"error":err.Error(),
		})
		return
	}
	allowed, err:=h.service.Check(
		c.Request.Context(),
		req.Key,
		req.RuleID,
	)
	if allowed {
        metrics.AllowedTotal.Inc()
    } else {
        metrics.RejectedTotal.Inc()
    }
	if err != nil {
		metrics.ErrorsTotal.Inc()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.CheckResponse{
		Allowed: allowed,
	})
}
