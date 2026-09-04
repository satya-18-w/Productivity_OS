package account_test

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// twoAccounts registers A and B against the same server and returns a client for each.
func twoAccounts(t *testing.T) (a, b *authClient) {
	t.Helper()
	a, _ = newFlow(t, time.Hour)
	a.register("a@iso.test", "Passw0rd!", "Asia/Kolkata")

	jar, _ := cookiejar.New(nil)
	b = &authClient{t: t, base: a.base, hc: &http.Client{Jar: jar}}
	b.register("b@iso.test", "Passw0rd!", "America/New_York")
	return a, b
}

func (c *authClient) timezone(t *testing.T) string {
	t.Helper()
	return readBody(t, c.do(http.MethodGet, "/api/account", ""))["timezone"].(string)
}

func TestIsolation_ReadIsAlwaysCallerOwn(t *testing.T) {
	a, b := twoAccounts(t)

	for _, q := range []string{"", "?account_id=a@iso.test", "?id=1", "?email=a@iso.test"} {
		resp := b.do(http.MethodGet, "/api/account"+q, "")
		require.Equal(t, "b@iso.test", readBody(t, resp)["email"], q)
	}
	require.Equal(t, "Asia/Kolkata", a.timezone(t))
}

func TestIsolation_AccountIDInBodyIsRejected(t *testing.T) {
	_, b := twoAccounts(t)
	// DecodeJSON disallows unknown fields, so a smuggled account_id is a 400,
	// never an override.
	resp := b.do(http.MethodPut, "/api/account/timezone",
		`{"timezone":"Europe/Paris","account_id":"a@iso.test"}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestIsolation_AccountIDHeaderIsIgnored(t *testing.T) {
	_, b := twoAccounts(t)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, b.base+"/api/account", nil)
	for _, ck := range b.hc.Jar.Cookies(req.URL) {
		req.AddCookie(ck)
	}
	req.Header.Set("X-Account-Id", "a@iso.test")
	req.Header.Set("Account-Id", "a@iso.test")
	resp, err := b.hc.Do(req)
	require.NoError(t, err)
	require.Equal(t, "b@iso.test", readBody(t, resp)["email"])
}

func TestIsolation_TimezoneChangeDoesNotCross(t *testing.T) {
	a, b := twoAccounts(t)

	resp := b.do(http.MethodPut, "/api/account/timezone", `{"timezone":"Europe/Paris"}`)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	require.Equal(t, "Europe/Paris", b.timezone(t))
	require.Equal(t, "Asia/Kolkata", a.timezone(t), "A's timezone is untouched")
}

func TestIsolation_PasswordChangeDoesNotCross(t *testing.T) {
	a, b := twoAccounts(t)

	resp := b.do(http.MethodPut, "/api/account/password",
		`{"current_password":"Passw0rd!","new_password":"NewPassw0rd!"}`)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// A's session still works.
	require.Equal(t, http.StatusOK, a.do(http.MethodGet, "/api/account", "").StatusCode)
	// B must re-login.
	require.Equal(t, http.StatusUnauthorized, b.do(http.MethodGet, "/api/account", "").StatusCode)
	// B's new password does not authenticate as A.
	login := b.do(http.MethodPost, "/api/sessions", `{"email":"a@iso.test","password":"NewPassw0rd!"}`)
	require.Equal(t, http.StatusUnauthorized, login.StatusCode)
}

func TestIsolation_LogoutOnlyEndsCallerSession(t *testing.T) {
	a, b := twoAccounts(t)

	resp := b.do(http.MethodDelete, "/api/sessions/current", "")
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, b.do(http.MethodGet, "/api/account", "").StatusCode)
	require.Equal(t, http.StatusOK, a.do(http.MethodGet, "/api/account", "").StatusCode)
}

func TestIsolation_SQLInjectionShapedValuesAreInert(t *testing.T) {
	c, _ := newFlow(t, time.Hour)

	evilPass := `x' OR '1'='1 -- plus padding to reach the minimum length`

	reg := c.do(http.MethodPost, "/api/accounts",
		`{"email":"inj@example.com","password":"`+evilPass+`","timezone":"UTC"}`)
	require.Equal(t, http.StatusCreated, reg.StatusCode)
	reg.Body.Close()

	// The table still exists and login works only with the literal password.
	ok := c.do(http.MethodPost, "/api/sessions", `{"email":"inj@example.com","password":"`+evilPass+`"}`)
	require.Equal(t, http.StatusOK, ok.StatusCode)
	ok.Body.Close()

	wrong := c.do(http.MethodPost, "/api/sessions", `{"email":"inj@example.com","password":"x' OR '1'='1"}`)
	require.Equal(t, http.StatusUnauthorized, wrong.StatusCode)
	wrong.Body.Close()
}
