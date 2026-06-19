package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

const requestsPerMinute = 100

// RateLimit — 100 req/min per user via Redis INCR.
func RateLimit(rdb *redis.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := fmt.Sprintf("ratelimit:%v:%d", c.Get(ctxUserID), time.Now().Unix()/60)

			count, err := rdb.Incr(context.Background(), key).Result()
			if err != nil {
				return next(c)
			}
			if count == 1 {
				rdb.Expire(context.Background(), key, time.Minute) //nolint:errcheck
			}
			if count > requestsPerMinute {
				slog.Warn("rate limit exceeded", "userID", c.Get(ctxUserID), "path", c.Request().URL.Path)
				return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
			}
			return next(c)
		}
	}
}
