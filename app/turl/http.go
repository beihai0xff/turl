package turl

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/beiai0xff/turl/configs"
)

// NewServer creates a new HTTP server.
func NewServer(c *configs.ServerConfig) (*http.Server, error) {
	h, err := newHandler(c)
	if err != nil {
		return nil, err
	}

	r := gin.Default()
	r.GET("/:short", h.Redirect)
	api := r.Group("/api")
	api.POST("/shorten", h.Create)

	//nolint:mnd
	return &http.Server{
		Addr:              fmt.Sprintf("%s:%d", c.Listen, c.Port),
		Handler:           http.TimeoutHandler(r.Handler(), 10*time.Second, "request timeout"),
		ReadHeaderTimeout: 500 * time.Millisecond,
		ReadTimeout:       500 * time.Millisecond,
	}, nil
}
