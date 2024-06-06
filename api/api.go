package api

const (
	// DefaultHost default host
	DefaultHost = "localhost"
	// DefaultPort default http server port
	DefaultPort = 10414

	// DefaultServerAddr default server addr
	// make sure the value is the same as DefaultHost:DefaultPort
	// we can't use the fmt.Sprintf("%s:%d", DefaultHost, DefaultPort) here
	// because the fmt.Sprintf will be executed at runtime, but the const value should be executed at compile time
	DefaultServerAddr = "localhost:10414"

	// VersionV1 api version
	VersionV1 = "/v1"
)
