package service

import (
	"context"
	"fmt"
	"time"

	"github.com/its-akshay/distributed-rate-limiter/internal/limiter"
	"github.com/its-akshay/distributed-rate-limiter/internal/repository"
)

type RateLimiterService struct {
	repo    repository.RuleRepositoryInterface
	limiter limiter.RateLimiter
}

func NewRateLimiterService(
	repo  repository.RuleRepositoryInterface,
	limiter limiter.RateLimiter,
) *RateLimiterService {
	return &RateLimiterService{
		repo:    repo,
		limiter: limiter,
	}
}

func (s *RateLimiterService) Check(
	ctx context.Context,
	key string,
	ruleId int64,
) (bool, error) {
	rule, err := s.repo.GetById(ctx, ruleId)
	if err != nil {
		return false, err
	}

	redisKey := fmt.Sprintf(
		"rate:%d:%s",
		ruleId,
		key,
	)

	return s.limiter.Allow(
		ctx, redisKey, rule.LimitCount, time.Duration(rule.WindowSeconds)*time.Second,
	)
}
