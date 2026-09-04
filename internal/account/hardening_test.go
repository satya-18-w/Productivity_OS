package account_test

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/account"
	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
)

// TestHardening_InternalErrorDoesNotLeak forces a database failure mid-request and
// asserts the client sees a generic INTERNAL error with no implementation detail
// (ADR-0002).
func TestHardening_InternalErrorDoesNotLeak(t *testing.T) {
	pool := pgtest.Pool(t)
	svc := account.NewService(pool, time.Hour)
	h := account.NewHandler(svc, false)
	mux := http.NewServeMux()
	h.MountPublic(mux)
	h.MountAuthed(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	jar, _ := cookiejar.New(nil)
	c := &authClient{t: t, base: srv.URL, hc: &http.Client{Jar: jar}}
	c.register("leak@example.com", "leak test password here", "UTC")

	pool.Close() // every subsequent query fails

	resp := c.do(http.MethodGet, "/api/account", "")
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	s := string(body)

	require.Contains(t, s, `"INTERNAL"`)
	for _, leak := range []string{"pool", "closed", "pgx", "SELECT", "sql", "password_hash", srv.URL} {
		require.NotContainsf(t, strings.ToLower(s), strings.ToLower(leak),
			"500 body must not contain %q", leak)
	}
}

func TestHardening_MalformedAndOversizedBodies(t *testing.T) {
	c, _ := newFlow(t, time.Hour)

	// malformed JSON -> 400, not 500
	resp := c.do(http.MethodPost, "/api/accounts", `{"email": `)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()

	// a password past the max length -> 400 VALIDATION_ERROR, not 500
	long := strings.Repeat("a", 500)
	resp = c.do(http.MethodPost, "/api/accounts",
		`{"email":"long@example.com","password":"`+long+`","timezone":"UTC"}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "VALIDATION_ERROR", readBody(t, resp)["error"].(map[string]any)["code"])

	// a body over the 1 MiB limit -> 400
	huge := strings.Repeat("x", 1<<20+16)
	resp = c.do(http.MethodPost, "/api/accounts",
		`{"email":"huge@example.com","password":"`+huge+`","timezone":"UTC"}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}
