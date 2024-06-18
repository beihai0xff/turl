// Package api define the API of the tiny URL service.
package api

import "fmt"

const (
	// DefaultHost default host
	DefaultHost = "localhost"
	// DefaultPort default http server port
	DefaultPort = 80

	// VersionV1 api version
	VersionV1 = "/v1"
)

// DefaultServerAddr default server addr
// make sure the value is formatted as "host:port"
var DefaultServerAddr = fmt.Sprintf("%s:%d", DefaultHost, DefaultPort)
