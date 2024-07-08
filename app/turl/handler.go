package turl

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/beihai0xff/turl/app/turl/model"
	"github.com/beihai0xff/turl/configs"
	"github.com/beihai0xff/turl/pkg/mapping"
)

// Handler represents the request handler.
type Handler struct {
	domain string
	s      Service
}

// NewHandler creates a new Handler.
func NewHandler(c *configs.ServerConfig) (*Handler, error) {
	s, err := newService(c)
	if err != nil {
		return nil, err
	}

	return &Handler{
		s:      s,
		domain: c.Domain,
	}, nil
}

// Create creates a new short URL from the long URL.
func (h *Handler) Create(c *gin.Context) {
	var req model.ShortenRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, &model.ShortenResponse{TinyURL: model.TinyURL{LongURL: req.LongURL}, Error: err.Error()})
		return
	}

	record, err := h.s.Create(c, []byte(req.LongURL))
	if err != nil {
		c.JSON(http.StatusInternalServerError, &model.ShortenResponse{TinyURL: model.TinyURL{LongURL: req.LongURL}, Error: err.Error()})
		return
	}

	record.ShortURL = fmt.Sprintf("%s/%s", h.domain, record.ShortURL)

	c.JSON(http.StatusOK, &model.ShortenResponse{TinyURL: *record})
}

// Redirect redirects the short URL to the original long URL temporarily if the short URL exists.
func (h *Handler) Redirect(c *gin.Context) {
	short := []byte(c.Param("short"))
	if len(short) > 8 || len(short) < 6 {
		c.JSON(http.StatusBadRequest, &model.ShortenResponse{TinyURL: model.TinyURL{ShortURL: string(short)}, Error: "invalid short URL"})
		return
	}

	long, err := h.s.Retrieve(c, short)
	if err != nil {
		t := model.TinyURL{ShortURL: string(short)}
		if errors.Is(err, mapping.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, &model.ShortenResponse{TinyURL: t, Error: "invalid short URL"})
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, &model.ShortenResponse{TinyURL: t, Error: "short URL not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, &model.ShortenResponse{TinyURL: t, Error: err.Error()})
	}

	c.Redirect(http.StatusFound, string(long))
}

// GetShortenInfo returns the original long URL of the short URL.
func (h *Handler) GetShortenInfo(c *gin.Context) {
	var req model.ShortenRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, &model.ShortenResponse{TinyURL: model.TinyURL{LongURL: req.LongURL}, Error: err.Error()})
		return
	}

	record, err := h.s.GetByLong(c, []byte(req.LongURL))
	if err != nil {
		c.JSON(http.StatusInternalServerError, &model.ShortenResponse{Error: err.Error()})
		return
	}

	record.ShortURL = fmt.Sprintf("%s/%s", h.domain, record.ShortURL)

	c.JSON(http.StatusOK, &model.ShortenResponse{TinyURL: *record})
}

// Close closes the handler.
func (h *Handler) Close() error {
	return h.s.Close()
}
