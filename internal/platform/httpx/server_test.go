package httpx

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	require.NoError(t, l.Close())
	return fmt.Sprintf("127.0.0.1:%d", port)
}

// get performs a GET and returns the status, always draining and closing the body.
func get(url string) (int, error) {
	resp, err := http.Get(url) //nolint:gosec // test-only, fixed local URL
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}

func TestServer_ServesThenShutsDownCleanly(t *testing.T) {
	addr := freePort(t)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	srv := NewServer(addr, mux, 2*time.Second)
	ctx, cancel := context.WithCancel(context.Background())

	runErr := make(chan error, 1)
	go func() { runErr <- srv.Run(ctx) }()

	require.Eventually(t, func() bool {
		status, err := get("http://" + addr + "/healthz")
		return err == nil && status == http.StatusOK
	}, 2*time.Second, 20*time.Millisecond)

	cancel()
	require.NoError(t, <-runErr, "Run must return nil on a context-driven shutdown")

	_, err := get("http://" + addr + "/healthz")
	require.Error(t, err, "server must stop accepting connections after shutdown")
}

func TestServer_DrainsInFlightRequest(t *testing.T) {
	addr := freePort(t)
	released := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("GET /slow", func(w http.ResponseWriter, _ *http.Request) {
		<-released
		WriteJSON(w, http.StatusOK, map[string]string{"status": "done"})
	})

	srv := NewServer(addr, mux, 3*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- srv.Run(ctx) }()

	require.Eventually(t, func() bool {
		c, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err != nil {
			return false
		}
		return c.Close() == nil
	}, 2*time.Second, 20*time.Millisecond)

	statusCh := make(chan int, 1)
	errCh := make(chan error, 1)
	go func() {
		status, err := get("http://" + addr + "/slow")
		if err != nil {
			errCh <- err
			return
		}
		statusCh <- status
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()        // shutdown begins while /slow is in flight
	close(released) // let the handler finish

	select {
	case status := <-statusCh:
		require.Equal(t, http.StatusOK, status)
	case err := <-errCh:
		require.NoError(t, err)
	}
	require.NoError(t, <-done)
}
