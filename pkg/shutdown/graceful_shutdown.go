// Package shutdown provides a graceful shutdown mechanism.
package shutdown

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/beihai0xff/turl/app/turl"
)

// OptionFunc is a function that can be used to configure a graceful shutdown.
type OptionFunc func(ctx context.Context) error

// HandlerShutdown shutdown the handler.
func HandlerShutdown(handler *turl.Handler) OptionFunc {
	return func(_ context.Context) error {
		if err := handler.Close(); err != nil {
			slog.Error("failed to shutdown handler", slog.Any("Error", err))
			return err
		}

		slog.Info("handler stopped")

		return nil
	}
}

// HTTPServerShutdown shutdown the HTTP server.
func HTTPServerShutdown(httpServer *http.Server) OptionFunc {
	return func(ctx context.Context) error {
		if err := httpServer.Shutdown(ctx); err != nil {
			slog.Error("failed to shutdown HTTP server",
				slog.String("Address", httpServer.Addr),
				slog.Any("Error", err),
			)

			return err
		}

		slog.Info("HTTP server stopped", slog.String("Address", httpServer.Addr))

		return nil
	}
}

// GracefulShutdown gracefully shutdown the server.
// 1. tell the load balancer this node is offline, and stop sending new requests
// 2. set the healthcheck status to unhealthy
// 3. stop accepting new HTTP requests and wait for existing HTTP requests to finish
// 4. flushing any buffered log entries
func GracefulShutdown(ctx context.Context, opts ...OptionFunc) {
	for _, opt := range opts {
		if err := opt(ctx); err != nil {
			slog.Error("failed to shutdown gracefully")
		}
	}
}
