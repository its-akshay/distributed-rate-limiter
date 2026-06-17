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

/*

Issue with the fixed window approach
if my window is of 1 man say
12:00:00 to 12:01:00

if a request came at 12:00:55 then it will get accepted as well 
ans after lets say at 12:01:00 another request came since this is in another window it will also get accepted
but in reality in span of 3 seconds all these requests are accepted


*/