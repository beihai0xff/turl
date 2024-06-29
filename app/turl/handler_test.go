package turl

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/beihai0xff/turl/internal/tests/mocks"
	"github.com/beihai0xff/turl/pkg/mapping"
)

func TestHandler_Create(t *testing.T) {
	mockService := mocks.NewMockTURLService(t)
	h := &Handler{s: mockService, domain: "https://www.example.com"}

	router := gin.Default()
	router.POST("/create", h.Create)

	t.Run("CreateSuccess", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewBuffer([]byte(`{"long_url":"https://www.example.com"}`)))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		mockService.EXPECT().Create(mock.Anything, mock.Anything).Return([]byte("abc123"), nil).Times(1)

		router.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code)
		require.Equal(t, `{"short_url":"https://www.example.com/abc123","long_url":"https://www.example.com","error":""}`, resp.Body.String())
	})

	t.Run("CreateInvalidURL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/create", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Body = http.NoBody
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("CreateURLFailed", func(t *testing.T) {
		testErr := errors.New("test error")
		mockService.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, testErr).Times(1)

		req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewBuffer([]byte(`{"long_url":"https://www.example.com"}`)))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusInternalServerError, resp.Code)
	})
}

func TestHandler_Redirect(t *testing.T) {
	mockService := mocks.NewMockTURLService(t)
	h := &Handler{s: mockService}

	router := gin.Default()
	router.GET("/redirect/:short", h.Redirect)

	t.Run("RedirectExistingURL", func(t *testing.T) {
		mockService.EXPECT().Retrieve(mock.Anything, []byte("abc123")).Return([]byte("https://www.example.com"), nil).Times(1)

		req := httptest.NewRequest(http.MethodGet, "/redirect/abc123", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusFound, resp.Code)
		require.Equal(t, "https://www.example.com", resp.Header().Get("Location"))
	})

	t.Run("RedirectNonExistingURL", func(t *testing.T) {
		mockService.EXPECT().Retrieve(mock.Anything, []byte("abc321")).Return(nil, gorm.ErrRecordNotFound).Times(1)

		req := httptest.NewRequest(http.MethodGet, "/redirect/abc321", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("RedirectInvalidURL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/redirect/123456789", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusBadRequest, resp.Code)

		mockService.EXPECT().Retrieve(mock.Anything, []byte("0123456")).Return(nil, mapping.ErrorInvalidCharacter).Times(1)
		req = httptest.NewRequest(http.MethodGet, "/redirect/0123456", nil)
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusBadRequest, resp.Code)
	})
}
