package middleware

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.Use(Logger())

	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	req, err := http.NewRequest(http.MethodGet, "/ping", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	fmt.Println(buf.String())

	require.Contains(t, buf.String(), "INFO")
	require.Contains(t, buf.String(), "GET")
	require.Contains(t, buf.String(), "/ping")
	require.Contains(t, buf.String(), "status=200")
}

func TestLogger_Error(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.Use(Logger())

	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusBadRequest, "pong")
	})

	req, err := http.NewRequest(http.MethodGet, "/ping", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)

	fmt.Println(buf.String())

	require.Contains(t, buf.String(), "ERROR")
	require.Contains(t, buf.String(), "GET")
	require.Contains(t, buf.String(), "/ping")
	require.Contains(t, buf.String(), "status=400")
}

func TestRateLimiter(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.Use(RateLimiter(1, 1))

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	req, err := http.NewRequest(http.MethodGet, "/ping", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println(w.Body.String())
	require.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.Use(HealthCheck())

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusInternalServerError, "pong")
	})

	req, err := http.NewRequest(http.MethodGet, "/healthcheck", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, `{"status":"ok"}`, w.Body.String())
}
