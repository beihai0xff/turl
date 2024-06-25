// Package middleware provides a set of middleware for the Gin framework.
package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/beihai0xff/turl/pkg/workqueue"
)

// Logger returns a middleware that logs the request.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		end := time.Now()

		var logFunc func(msg string, args ...any)

		if c.Writer.Status() < http.StatusBadRequest {
			logFunc = slog.Info
		} else {
			logFunc = slog.Error
		}

		logFunc("http request",
			slog.String("ip", c.ClientIP()),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("proto", c.Request.Proto),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", end.Sub(start)),
			slog.String("user_agent", c.Request.UserAgent()),
			slog.String("error", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)
	}
}

// RateLimiter returns a middleware that limits the number of requests per second.
func RateLimiter(r, b int) gin.HandlerFunc {
	limiter := workqueue.NewBucketRateLimiter[any](rate.NewLimiter(rate.Limit(r), b))

	return func(c *gin.Context) {
		if !limiter.Take(c.Request.RemoteAddr) {
			c.String(http.StatusTooManyRequests, "rate limit exceeded, retry later")
		} else {
			c.Next()
		}
	}
}

// HealthCheck returns a middleware that checks the health of the server.
func HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/healthcheck" {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		} else {
			c.Next()
		}
	}
}
