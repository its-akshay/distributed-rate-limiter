package service

import "context"

type RateLimiterServiceInterface interface {
	Check(ctx context.Context,key string,ruleId int64) (bool, error)
}
