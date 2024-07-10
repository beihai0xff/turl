package turl

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/beihai0xff/turl/configs"
	"github.com/beihai0xff/turl/pkg/db/redis"
	"github.com/beihai0xff/turl/pkg/middleware"
	"github.com/beihai0xff/turl/pkg/workqueue"
)

// HealthCheckPath health check path
const HealthCheckPath = "/healthcheck"

// NewServer creates a new HTTP server.
func NewServer(h *Handler, c *configs.ServerConfig) (*http.Server, error) {
	router := gin.New()
	router.UseH2C = true

	if c.Debug {
		// register pprof router
		pprof.Register(router, "api/debug/pprof")
	}

	router.Use(middleware.Logger(), middleware.HealthCheck(HealthCheckPath))

	router.Use(gin.Recovery()) // recover from any panics, should be the last middleware
	router.GET("/:short", h.Redirect).Use(middleware.RateLimiter(
		workqueue.NewBucketRateLimiter[any](rate.NewLimiter(rate.Limit(c.StandAloneReadRate), c.StandAloneReadBurst))))

	if !c.Readonly {
		rdb := redis.Client(c.Cache.Redis)
		api := router.Group("/api").Use(middleware.RateLimiter(
			workqueue.NewItemRedisTokenRateLimiter[any](rdb, c.GlobalRateLimitKey, c.GlobalWriteRate, c.GlobalWriteBurst, time.Second)))
		api.POST("/shorten", h.Create)
		api.GET("/shorten", h.GetShortenInfo)
		api.DELETE("/shorten", h.Delete)
	}

	return &http.Server{
		Addr:              fmt.Sprintf("%s:%d", c.Listen, c.Port),
		Handler:           http.TimeoutHandler(router.Handler(), c.RequestTimeout, "request timeout"),
		ReadHeaderTimeout: c.RequestTimeout,
		ReadTimeout:       c.RequestTimeout,
		WriteTimeout:      c.RequestTimeout,
	}, nil
}
