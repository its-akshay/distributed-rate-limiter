package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/its-akshay/distributed-rate-limiter/internal/model"
	"github.com/its-akshay/distributed-rate-limiter/internal/repository"
)

type RuleHandler struct {
	repo *repository.RuleRepository
}

func NewRuleHandler(repo *repository.RuleRepository) *RuleHandler {
	return &RuleHandler{
		repo: repo,
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
