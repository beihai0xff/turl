package turl

// ShortenRequest is the request of shorten API
type ShortenRequest struct {
	// LongURL is the original long URL
	LongURL string `json:"long_url" form:"long_url" xml:"long_url" binding:"required,http_url"`
}

// ShortenResponse is the response of shorten API
type ShortenResponse struct {
	// ShortURL is the shortened URL
	ShortURL string `json:"short_url"`
	// LongURL is the original long URL
	LongURL string `json:"long_url"`
	// Error is the error message if any error occurs
	Error string `json:"error"`
}
