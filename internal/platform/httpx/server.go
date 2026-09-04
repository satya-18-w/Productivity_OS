// Package httpx holds the shared HTTP infrastructure: the server and its lifecycle,
// middleware, the error envelope, and JSON helpers. No business logic lives here.
package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Server wraps http.Server with a context-driven graceful shutdown.
type Server struct {
	http  *http.Server
	grace time.Duration
}

// NewServer builds a server bound to addr serving handler.
func NewServer(addr string, handler http.Handler, grace time.Duration) *Server {
	return &Server{
		http: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 10 * time.Second,
		},
		grace: grace,
	}
}

// Run serves until ctx is cancelled, then stops accepting connections and lets
// in-flight requests finish within the grace period. It returns nil on a clean
// shutdown.
func (s *Server) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.http.Addr)
	if err != nil {
		return err
	}

	serveErr := make(chan error, 1)
	go func() {
		slog.Info("server listening", slog.String("addr", s.http.Addr))
		serveErr <- s.http.Serve(ln)
	}()

	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		slog.Info("shutdown signal received", slog.Duration("grace", s.grace))
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.grace)
		defer cancel()
		return s.http.Shutdown(shutdownCtx)
	}
}
