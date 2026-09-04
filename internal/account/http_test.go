package account_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/account"
	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
)

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	pool := pgtest.Pool(t)
	svc := account.NewService(pool, time.Hour)
	h := account.NewHandler(svc, false)
	mux := http.NewServeMux()
	h.MountPublic(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func postJSON(t *testing.T, srv *httptest.Server, path, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost,
		srv.URL+path, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	return resp
}

func readBody(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()
	var m map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&m))
	return m
}

func sessionCookie(resp *http.Response) *http.Cookie {
	for _, c := range resp.Cookies() {
		if c.Name == account.SessionCookieName {
			return c
		}
	}
	return nil
}

func TestRegisterEndpoint(t *testing.T) {
	srv := newServer(t)

	resp := postJSON(t, srv, "/api/accounts",
		`{"email":"web@example.com","password":"a good long password","timezone":"Europe/London"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	body := readBody(t, resp)
	require.Equal(t, "web@example.com", body["email"])
	require.Equal(t, "Europe/London", body["timezone"])

	c := sessionCookie(resp)
	require.NotNil(t, c)
	require.True(t, c.HttpOnly)
	require.Equal(t, http.SameSiteLaxMode, c.SameSite)

	// duplicate
	dup := postJSON(t, srv, "/api/accounts",
		`{"email":"WEB@example.com","password":"different password","timezone":"UTC"}`)
	require.Equal(t, http.StatusConflict, dup.StatusCode)
	require.Equal(t, "EMAIL_ALREADY_REGISTERED", readBody(t, dup)["error"].(map[string]any)["code"])
}

func TestRegisterEndpoint_Validation(t *testing.T) {
	srv := newServer(t)
	resp := postJSON(t, srv, "/api/accounts",
		`{"email":"bad","password":"short","timezone":"Mars/Base"}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	errObj := readBody(t, resp)["error"].(map[string]any)
	require.Equal(t, "VALIDATION_ERROR", errObj["code"])
	fields := errObj["fields"].(map[string]any)
	require.Contains(t, fields, "email")
	require.Contains(t, fields, "password")
	require.Contains(t, fields, "timezone")
}

func TestRegisterEndpoint_MalformedJSON(t *testing.T) {
	srv := newServer(t)
	resp := postJSON(t, srv, "/api/accounts", `{"email":`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	_ = readBody(t, resp)
}

func TestLoginEndpoint_NoUserEnumeration(t *testing.T) {
	srv := newServer(t)
	postJSON(t, srv, "/api/accounts",
		`{"email":"real@example.com","password":"the actual password","timezone":"UTC"}`).Body.Close()

	unknown := postJSON(t, srv, "/api/sessions", `{"email":"ghost@example.com","password":"the actual password"}`)
	wrong := postJSON(t, srv, "/api/sessions", `{"email":"real@example.com","password":"not the password"}`)

	require.Equal(t, http.StatusUnauthorized, unknown.StatusCode)
	require.Equal(t, http.StatusUnauthorized, wrong.StatusCode)

	defer unknown.Body.Close()
	defer wrong.Body.Close()
	ub, _ := io.ReadAll(unknown.Body)
	wb, _ := io.ReadAll(wrong.Body)
	require.Equal(t, string(wb), string(ub), "identical bodies for unknown email and wrong password")
}

func TestLoginEndpoint_Success(t *testing.T) {
	srv := newServer(t)
	postJSON(t, srv, "/api/accounts",
		`{"email":"ok@example.com","password":"login works fine","timezone":"UTC"}`).Body.Close()

	resp := postJSON(t, srv, "/api/sessions", `{"email":"ok@example.com","password":"login works fine"}`)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	require.NotNil(t, sessionCookie(resp))
}

func TestLoginEndpoint_RateLimited(t *testing.T) {
	srv := newServer(t)
	postJSON(t, srv, "/api/accounts",
		`{"email":"rl@example.com","password":"rate limit me now","timezone":"UTC"}`).Body.Close()

	body := `{"email":"rl@example.com","password":"wrong every time"}`
	for i := 0; i < 5; i++ {
		r := postJSON(t, srv, "/api/sessions", body)
		require.Equal(t, http.StatusUnauthorized, r.StatusCode)
		r.Body.Close()
	}
	sixth := postJSON(t, srv, "/api/sessions", body)
	require.Equal(t, http.StatusTooManyRequests, sixth.StatusCode)
	require.Equal(t, "RATE_LIMITED", readBody(t, sixth)["error"].(map[string]any)["code"])
}
