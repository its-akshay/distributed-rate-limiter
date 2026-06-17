package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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

func (h *RuleHandler) Check(c *gin.Context) {
	var req model.CheckRequest

	if err:=c.ShouldBindJSON(&req); err!=nil {
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
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.CheckResponse{
		Allowed: allowed,
	})
}
