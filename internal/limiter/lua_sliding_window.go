package limiter

import (
	"context"
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed scripts/sliding_window.lua
var slidingWindowScript string

type LuaSlidingWindowLimiter struct {
	rdb    *redis.Client
	script *redis.Script
}

func NewLuaSlidingWindowLimiter(rdb *redis.Client) *LuaSlidingWindowLimiter {
	return &LuaSlidingWindowLimiter{
		rdb:    rdb,
		script: redis.NewScript(slidingWindowScript),
	}
}

func (l *LuaSlidingWindowLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {

	now := time.Now().UnixMilli()
	res, err := l.script.Run(
		ctx,
		l.rdb,
		[]string{key},
		now,
		window.Milliseconds(),
		limit,
	).Int()

	if err != nil {
		return false, err
	}
	return res == 1, nil

}
