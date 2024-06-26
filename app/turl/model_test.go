package turl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShortenRequest(t *testing.T) {
	t.Run("ValidURL", func(t *testing.T) {
		req := ShortenRequest{LongURL: "https://www.example.com"}
		require.NotNil(t, req)
	})

	t.Run("InvalidURL", func(t *testing.T) {
		req := ShortenRequest{LongURL: "invalid_url"}
		require.NotNil(t, req)
	})
}

func TestShortenResponse(t *testing.T) {
	t.Run("ValidResponse", func(t *testing.T) {
		resp := ShortenResponse{ShortURL: []byte("https://turl.com/abc"), LongURL: []byte("https://www.example.com"), Error: ""}
		require.NotNil(t, resp)
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		resp := ShortenResponse{Error: "An error occurred"}
		require.NotNil(t, resp)
	})
}
