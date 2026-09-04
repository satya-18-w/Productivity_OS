package account

import (
	"crypto/subtle"
	"errors"
	"net/http"

	"github.com/satya-18-w/productivity-os/internal/platform/httpx"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
)

// csrfCookieName is the readable double-submit CSRF cookie. The SPA echoes its
// value in the X-CSRF-Token header on state-changing requests (ADR-0004).
const csrfCookieName = "csrf_token"

// csrfHeaderName is where the SPA sends the token back.
const csrfHeaderName = "X-CSRF-Token"

// RequireAuth resolves the session cookie to an account and puts its identity in
// the request context. Requests without a valid session get 401 UNAUTHENTICATED.
// The account identity is never taken from anywhere else (ADR-0004, v1.md N3).
func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(SessionCookieName)
		if err != nil || cookie.Value == "" {
			unauthenticated(w, r)
			return
		}
		id, err := h.svc.ResolveSession(r.Context(), cookie.Value)
		if err != nil {
			if errors.Is(err, ErrSessionInvalid) {
				unauthenticated(w, r)
				return
			}
			httpx.WriteError(w, r, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(reqctx.WithIdentity(r.Context(), id)))
	})
}

// RequireCSRF enforces the double-submit token on unsafe methods: the
// X-CSRF-Token header must be present and equal to the csrf_token cookie.
func (h *Handler) RequireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie(csrfCookieName)
		header := r.Header.Get(csrfHeaderName)
		if err != nil || cookie.Value == "" || header == "" ||
			subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(header)) != 1 {
			httpx.WriteError(w, r, httpx.NewError(http.StatusForbidden, httpx.CodeForbidden,
				"CSRF token missing or invalid"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func unauthenticated(w http.ResponseWriter, r *http.Request) {
	httpx.WriteError(w, r, httpx.NewError(http.StatusUnauthorized, httpx.CodeUnauthenticated,
		"Authentication required"))
}

func (h *Handler) setCSRFCookie(w http.ResponseWriter, token string) {
	//nolint:gosec // G124: intentionally readable by JS (double-submit); Secure is env-conditional
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearCSRFCookie(w http.ResponseWriter) {
	//nolint:gosec // G124: see setCSRFCookie
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
