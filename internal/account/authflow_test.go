package account_test

import (
	"context"
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

// authClient is an HTTP client with a cookie jar that also sends the double-submit
// CSRF header, mimicking the SPA.
type authClient struct {
	t    *testing.T
	base string
	hc   *http.Client
}

func newFlow(t *testing.T, ttl time.Duration) (*authClient, *httptest.Server) {
	t.Helper()
	pool := pgtest.Pool(t)
	svc := account.NewService(pool, ttl)
	h := account.NewHandler(svc, false)
	mux := http.NewServeMux()
	h.MountPublic(mux)
	h.MountAuthed(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	return &authClient{t: t, base: srv.URL, hc: &http.Client{Jar: jar}}, srv
}

func (c *authClient) do(method, path, body string) *http.Response {
	c.t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, c.base+path, strings.NewReader(body))
	require.NoError(c.t, err)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, ck := range c.hc.Jar.Cookies(req.URL) {
		if ck.Name == "csrf_token" {
			req.Header.Set("X-CSRF-Token", ck.Value)
		}
	}
	resp, err := c.hc.Do(req)
	require.NoError(c.t, err)
	return resp
}

func (c *authClient) register(email, pw, tz string) {
	c.t.Helper()
	resp := c.do(http.MethodPost, "/api/accounts",
		`{"email":"`+email+`","password":"`+pw+`","timezone":"`+tz+`"}`)
	require.Equal(c.t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()
}

func TestAuthFlow_RequireAuth(t *testing.T) {
	c, _ := newFlow(t, time.Hour)

	// no session
	resp := c.do(http.MethodGet, "/api/account", "")
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	require.Equal(t, "UNAUTHENTICATED", readBody(t, resp)["error"].(map[string]any)["code"])

	c.register("flow@example.com", "the flow password", "UTC")

	resp = c.do(http.MethodGet, "/api/account", "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "flow@example.com", readBody(t, resp)["email"])
}

func TestAuthFlow_BadCookie(t *testing.T) {
	_, srv := newFlow(t, time.Hour)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/account", nil)
	req.AddCookie(&http.Cookie{Name: account.SessionCookieName, Value: "garbage-token"})
	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthFlow_CSRFRequiredForWrites(t *testing.T) {
	c, srv := newFlow(t, time.Hour)
	c.register("csrf@example.com", "csrf test password", "UTC")

	// A write without the CSRF header is rejected even with a valid session.
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut, srv.URL+"/api/account/timezone",
		strings.NewReader(`{"timezone":"Asia/Kolkata"}`))
	req.Header.Set("Content-Type", "application/json")
	for _, ck := range c.hc.Jar.Cookies(req.URL) {
		req.AddCookie(ck) // session + csrf cookie, but no X-CSRF-Token header
	}
	resp, err := c.hc.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()

	// With the header it succeeds.
	ok := c.do(http.MethodPut, "/api/account/timezone", `{"timezone":"Asia/Kolkata"}`)
	require.Equal(t, http.StatusNoContent, ok.StatusCode)
	ok.Body.Close()

	got := c.do(http.MethodGet, "/api/account", "")
	require.Equal(t, "Asia/Kolkata", readBody(t, got)["timezone"])
}

func TestAuthFlow_Logout(t *testing.T) {
	c, _ := newFlow(t, time.Hour)
	c.register("bye@example.com", "logout me please", "UTC")

	resp := c.do(http.MethodDelete, "/api/sessions/current", "")
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	after := c.do(http.MethodGet, "/api/account", "")
	require.Equal(t, http.StatusUnauthorized, after.StatusCode)
}

func TestAuthFlow_PasswordChangeEndsSession(t *testing.T) {
	c, _ := newFlow(t, time.Hour)
	c.register("pw@example.com", "original secret pw", "UTC")

	resp := c.do(http.MethodPut, "/api/account/password",
		`{"current_password":"original secret pw","new_password":"a fresh new secret"}`)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// current session is dead
	after := c.do(http.MethodGet, "/api/account", "")
	require.Equal(t, http.StatusUnauthorized, after.StatusCode)

	// wrong current password
	c.register("pw2@example.com", "another original", "UTC")
	bad := c.do(http.MethodPut, "/api/account/password",
		`{"current_password":"not it","new_password":"whatever long enough"}`)
	require.Equal(t, http.StatusUnauthorized, bad.StatusCode)
	bad.Body.Close()
}

func TestAuthFlow_CrossAccountIsolation(t *testing.T) {
	a, srv := newFlow(t, time.Hour)
	a.register("a@example.com", "account a password", "Asia/Kolkata")

	jarB, _ := cookiejar.New(nil)
	b := &authClient{t: t, base: srv.URL, hc: &http.Client{Jar: jarB}}
	b.register("b@example.com", "account b password", "America/New_York")

	// B reads only B's data, regardless of any body/query.
	resp := b.do(http.MethodGet, "/api/account?account_id=whatever", "")
	require.Equal(t, "b@example.com", readBody(t, resp)["email"])

	// B changes B's timezone; A is untouched.
	resp = b.do(http.MethodPut, "/api/account/timezone", `{"timezone":"Europe/Paris"}`)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	resp = a.do(http.MethodGet, "/api/account", "")
	require.Equal(t, "Asia/Kolkata", readBody(t, resp)["timezone"])
}

func TestAuthFlow_ExpiredSession(t *testing.T) {
	c, _ := newFlow(t, 30*time.Millisecond)
	c.register("stale@example.com", "expire this session", "UTC")
	time.Sleep(60 * time.Millisecond)

	resp := c.do(http.MethodGet, "/api/account", "")
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
