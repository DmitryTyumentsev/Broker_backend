package middleware

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"Broker_backend/services/apigateway/internal/config"
	"Broker_backend/services/apigateway/internal/transport/http/httperr"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var redisRateLimitScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
	redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
local ttl = redis.call("PTTL", KEYS[1])
return {current, ttl}
`)

func RedisRateLimit(client *redis.Client, cfg config.RateLimitConfig, logger *zap.Logger) fiber.Handler {
	if logger == nil {
		logger = zap.NewNop()
	}

	return func(c *fiber.Ctx) error {
		if !cfg.Enabled {
			return c.Next()
		}

		if client == nil {
			return httperr.WriteServiceUnavailable(c, "rate limiter is not configured")
		}

		ctx := c.UserContext()
		if ctx == nil {
			ctx = context.Background()
		}

		key := rateLimitKey(cfg.Prefix, c)
		result, err := incrementRateLimit(ctx, client, key, cfg.Window)
		if err != nil {
			logger.Warn("rate limit failed", zap.Error(err), zap.String("key", key))
			return httperr.WriteServiceUnavailable(c, "rate limiter unavailable")
		}

		remaining := cfg.Limit - result.Count
		if remaining < 0 {
			remaining = 0
		}

		c.Set("X-RateLimit-Limit", strconv.FormatInt(cfg.Limit, 10))
		c.Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Set("X-RateLimit-Window", cfg.Window.String())

		if result.Count > cfg.Limit {
			if result.TTL > 0 {
				c.Set("Retry-After", strconv.FormatInt(retryAfterSeconds(result.TTL), 10))
			}

			return httperr.WriteTooManyRequests(c, "rate limit exceeded")
		}

		return c.Next()
	}
}

type rateLimitResult struct {
	Count int64
	TTL   time.Duration
}

func incrementRateLimit(
	ctx context.Context,
	client *redis.Client,
	key string,
	window time.Duration,
) (rateLimitResult, error) {
	windowMillis := window.Milliseconds()
	if windowMillis <= 0 {
		windowMillis = 1
	}

	raw, err := redisRateLimitScript.Run(ctx, client, []string{key}, windowMillis).Result()
	if err != nil {
		return rateLimitResult{}, err
	}

	values, ok := raw.([]interface{})
	if !ok || len(values) != 2 {
		return rateLimitResult{}, fmt.Errorf("unexpected redis script result: %T", raw)
	}

	count, err := redisResultInt64(values[0])
	if err != nil {
		return rateLimitResult{}, fmt.Errorf("parse count: %w", err)
	}

	ttlMillis, err := redisResultInt64(values[1])
	if err != nil {
		return rateLimitResult{}, fmt.Errorf("parse ttl: %w", err)
	}

	if ttlMillis < 0 {
		ttlMillis = 0
	}

	return rateLimitResult{
		Count: count,
		TTL:   time.Duration(ttlMillis) * time.Millisecond,
	}, nil
}

func redisResultInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	case []byte:
		return strconv.ParseInt(string(v), 10, 64)
	default:
		return 0, fmt.Errorf("unexpected integer type %T", value)
	}
}

func retryAfterSeconds(ttl time.Duration) int64 {
	if ttl <= 0 {
		return 1
	}

	seconds := int64((ttl + time.Second - time.Nanosecond) / time.Second)
	if seconds <= 0 {
		return 1
	}

	return seconds
}

func rateLimitKey(prefix string, c *fiber.Ctx) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "rl"
	}

	identity := strings.TrimSpace(c.IP())
	if identity == "" {
		identity = "unknown"
	}

	return fmt.Sprintf("%s:%s:%s:%s", prefix, identity, c.Method(), c.Path())
}
