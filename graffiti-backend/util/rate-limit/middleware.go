package rate_limit

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type TokenBucketLimiter struct {
	RedisClient *redis.Client
	Capacity    int
	RefillRate  float64
	Window      time.Duration
}

func NewTokenBucketLimiter(redisAddr string, capacity int, refillRate float64, ttl time.Duration) *TokenBucketLimiter {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return &TokenBucketLimiter{
		RedisClient: rdb,
		Capacity:    capacity,
		RefillRate:  refillRate,
		Window:      ttl,
	}
}
func (tb *TokenBucketLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var identifier string

		if user, exists := c.Get("currentUser"); exists {
			identifier = fmt.Sprintf("user:%v", user.(interface{ GetID() string }).GetID())
		} else {
			identifier = fmt.Sprintf("ip:%s", c.ClientIP())
		}

		ctx := context.Background()
		now := float64(time.Now().Unix())
		keyTokens := fmt.Sprintf("tokens:%s", identifier)
		keyLastRefill := fmt.Sprintf("tokens_last_refill:%s", identifier)

		luaScript := redis.NewScript(`
			local tokens_key = KEYS[1]
			local last_refill_key = KEYS[2]
			local now = tonumber(ARGV[1])
			local capacity = tonumber(ARGV[2])
			local refill_rate = tonumber(ARGV[3])
			local ttl = tonumber(ARGV[4])

			local tokens = tonumber(redis.call("get", tokens_key)) or capacity
			local last_refill = tonumber(redis.call("get", last_refill_key)) or now

			local elapsed = now - last_refill
			local refill = elapsed * refill_rate
			tokens = math.min(capacity, tokens + refill)

			if tokens < 1 then
				return {0, tokens}
			end

			tokens = tokens - 1
			redis.call("setex", tokens_key, ttl, tokens)
			redis.call("setex", last_refill_key, ttl, now)
			return {1, tokens}
		`)

		result, err := luaScript.Run(ctx, tb.RedisClient, []string{keyTokens, keyLastRefill},
			now, 1, 1.0, int(tb.Window.Seconds())).Result() // Capacity=1, RefillRate=1.0/s

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Rate limiter error"})
			return
		}

		res := result.([]interface{})
		allowed := res[0].(int64)
		remaining := res[1].(float64)

		if allowed == 0 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message":          "Rate limit exceeded. Please wait.",
				"tokens_remaining": remaining,
			})
			return
		}

		c.Next()
	}
}
