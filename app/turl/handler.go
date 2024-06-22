package turl

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/beiai0xff/turl/configs"
	"github.com/beiai0xff/turl/pkg/mapping"
)

type handler struct {
	t *TinyURL
}

func newHandler(c *configs.ServerConfig) (*handler, error) {
	t, err := NewTinyURL(c)
	if err != nil {
		return nil, err
	}

	return &handler{
		t: t,
	}, nil
}

// Create creates a new short URL from the long URL.
func (h *handler) Create(c *gin.Context) {
	var req ShortenRequest

	if c.ShouldBind(&req) != nil {
		c.JSON(http.StatusBadRequest, &ShortenResponse{LongURL: req.LongURL,
			Error: "invalid request param 'long_url'"})
		return
	}

	short, err := h.t.Create(c, req.LongURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &ShortenResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, &ShortenResponse{ShortURL: short, LongURL: req.LongURL})
}

// Redirect redirects the short URL to the original long URL temporarily if the short URL exists.
func (h *handler) Redirect(c *gin.Context) {
	short := []byte(c.Param("short"))
	if len(short) > 8 || len(short) < 6 {
		c.JSON(http.StatusBadRequest, &ShortenResponse{ShortURL: short, Error: "invalid short URL"})
		return
	}

	long, err := h.t.Retrieve(c, short)
	if err != nil {
		if errors.Is(err, mapping.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, &ShortenResponse{ShortURL: short, Error: "invalid short URL"})
			return
		}

		c.JSON(http.StatusInternalServerError, &ShortenResponse{ShortURL: short, Error: err.Error()})
	}

	c.Redirect(http.StatusFound, string(long))
}
