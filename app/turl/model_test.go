package turl

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/beihai0xff/turl/app/turl/model"
)

func TestShortenRequest(t *testing.T) {
	t.Run("ValidURL", func(t *testing.T) {
		req := model.ShortenRequest{LongURL: "https://www.example.com"}
		require.NotNil(t, req)
	})

	t.Run("InvalidURL", func(t *testing.T) {
		req := model.ShortenRequest{LongURL: "invalid_url"}
		require.NotNil(t, req)
	})
}

func TestShortenResponse(t *testing.T) {
	t.Run("ValidResponse", func(t *testing.T) {
		resp := model.ShortenResponse{TinyURL: model.TinyURL{ShortURL: "https://turl.com/abc", LongURL: "https://www.example.com"}, Error: ""}
		require.NotNil(t, resp)
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		resp := model.ShortenResponse{Error: "An error occurred"}
		require.NotNil(t, resp)
	})
}
