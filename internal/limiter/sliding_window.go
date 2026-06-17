package limiter

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type SlidingWindowLimiter struct {
	rdb *redis.Client
}

func NewSlidingWindowLimiter(
	rdb *redis.Client,
) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		rdb: rdb,
	}
}

func (l *SlidingWindowLimiter) Allow(
	ctx context.Context,
	key string,
	limit int,
	window time.Duration,
) (bool, error) {
	now := time.Now().UnixMilli()
	windowStart := now - window.Milliseconds()

	//redis sorted set
	_, err := l.rdb.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10)).Result()
	if err != nil {
		return false, err
	}
	count, err := l.rdb.ZCard(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count >= int64(limit) {
		return false, nil
	}

	err = l.rdb.ZAdd(ctx, key, redis.Z{
		Score:  float64(now),
		Member: now,
	}).Err()
	if err != nil {
		return false, err
	}

	err = l.rdb.Expire(ctx, key, window).Err()
	if err != nil {
		return false, err
	}

	return true, nil
}

//zadd -> add element to a sorted set with a score
//zcard -> return the number of elements in sorted sets
//zremrangebyscore ->remove all elements whose score lient in the given range
//expire set TTL

//it gives us dynamically last N seconds window
