package turl

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/beihai0xff/turl/configs"
	"github.com/beihai0xff/turl/pkg/middleware"
)

// NewServer creates a new HTTP server.
func NewServer(h *Handler, c *configs.ServerConfig) (*http.Server, error) {
	router := gin.New()
	router.Use(middleware.Logger(), middleware.HealthCheck(), middleware.RateLimiter(c.Rate, c.Burst))

	router.Use(gin.Recovery()) // recover from any panics, should be the last middleware
	router.GET("/:short", h.Redirect)
	api := router.Group("/api")
	api.POST("/shorten", h.Create)

	//nolint:mnd
	return &http.Server{
		Addr:              fmt.Sprintf("%s:%d", c.Listen, c.Port),
		Handler:           http.TimeoutHandler(router.Handler(), 10*time.Second, "request timeout"),
		ReadHeaderTimeout: 500 * time.Millisecond,
		ReadTimeout:       500 * time.Millisecond,
	}, nil
}
