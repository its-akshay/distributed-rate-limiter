package limiter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type FixedWindowLimiter struct {
	rdb *redis.Client
}

func NewFixedWindowLimiter (
	rdb *redis.Client,
) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		rdb: rdb,
	}
}

func (l* FixedWindowLimiter) Allow(
	ctx context.Context,
	key string,
	limit int, 
	window time.Duration,
) (bool , error ){
	// if key isnt set it create with value one or else increment by 1
	//otherwise increase by one 
	count, err := l.rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count ==1 {
		err = l.rdb.Expire(ctx, key, window).Err()
		if err != nil {
			return false, err
		}
	}
	return count <= int64(limit), nil
}