package turl

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/beihai0xff/turl/configs"
	"github.com/beihai0xff/turl/pkg/mapping"
)

// Handler represents the request handler.
type Handler struct {
	s Service
}

// NewHandler creates a new Handler.
func NewHandler(c *configs.ServerConfig) (*Handler, error) {
	s, err := newTinyURLService(c)
	if err != nil {
		return nil, err
	}

	return &Handler{
		s: s,
	}, nil
}

// Create creates a new short URL from the long URL.
func (h *Handler) Create(c *gin.Context) {
	var req ShortenRequest

	if c.ShouldBind(&req) != nil {
		c.JSON(http.StatusBadRequest, &ShortenResponse{LongURL: req.LongURL,
			Error: "invalid request param 'long_url'"})
		return
	}

	short, err := h.s.Create(c, []byte(req.LongURL))
	if err != nil {
		c.JSON(http.StatusInternalServerError, &ShortenResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, &ShortenResponse{ShortURL: string(short), LongURL: req.LongURL})
}

// Redirect redirects the short URL to the original long URL temporarily if the short URL exists.
func (h *Handler) Redirect(c *gin.Context) {
	short := []byte(c.Param("short"))
	if len(short) > 8 || len(short) < 6 {
		c.JSON(http.StatusBadRequest, &ShortenResponse{ShortURL: string(short), Error: "invalid short URL"})
		return
	}

	long, err := h.s.Retrieve(c, short)
	if err != nil {
		if errors.Is(err, mapping.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, &ShortenResponse{ShortURL: string(short), Error: "invalid short URL"})
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, &ShortenResponse{ShortURL: string(short), Error: "short URL not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, &ShortenResponse{ShortURL: string(short), Error: err.Error()})
	}

	c.Redirect(http.StatusFound, string(long))
}

// Close closes the handler.
func (h *Handler) Close() error {
	return h.s.Close()
}
