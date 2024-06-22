package turl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShortenRequest(t *testing.T) {
	t.Run("ValidURL", func(t *testing.T) {
		req := ShortenRequest{LongURL: []byte("https://www.example.com")}
		assert.NotNil(t, req)
	})

	t.Run("InvalidURL", func(t *testing.T) {
		req := ShortenRequest{LongURL: []byte("invalid_url")}
		assert.NotNil(t, req)
	})
}

func TestShortenResponse(t *testing.T) {
	t.Run("ValidResponse", func(t *testing.T) {
		resp := ShortenResponse{ShortURL: []byte("https://turl.com/abc"), LongURL: []byte("https://www.example.com"), Error: ""}
		assert.NotNil(t, resp)
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		resp := ShortenResponse{Error: "An error occurred"}
		assert.NotNil(t, resp)
	})
}
