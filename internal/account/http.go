package account

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/satya-18-w/productivity-os/internal/platform/httpx"
	"github.com/satya-18-w/productivity-os/internal/platform/ratelimit"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
)

// SessionCookieName is the name of the session cookie.
const SessionCookieName = "session"

// Handler serves the account and session HTTP endpoints.
type Handler struct {
	svc           Service
	secure        bool
	loginLimiter  *ratelimit.Limiter // failed logins per email+IP
	authIPLimiter *ratelimit.Limiter // all auth requests per IP (bounds Argon2id cost)
}

// NewHandler builds the account handler. secure controls the cookie Secure flag
// (true in production).
func NewHandler(svc Service, secure bool) *Handler {
	return &Handler{
		svc:           svc,
		secure:        secure,
		loginLimiter:  ratelimit.New(5, 15*time.Minute),
		authIPLimiter: ratelimit.New(30, 5*time.Minute),
	}
}

// throttleByIP records an auth request against the per-IP budget and reports
// whether it is allowed. It bounds unauthenticated Argon2id work.
//
// NOTE: clientIP reads RemoteAddr. Once the deployment/proxy shape is fixed
// (ADR-0008), revisit whether to trust a forwarded-for header here.
func (h *Handler) throttleByIP(w http.ResponseWriter, r *http.Request) bool {
	ip := clientIP(r)
	if !h.authIPLimiter.Allowed(ip) {
		httpx.WriteError(w, r, httpx.NewError(http.StatusTooManyRequests, httpx.CodeRateLimited,
			"Too many requests. Try again later."))
		return false
	}
	h.authIPLimiter.Fail(ip)
	return true
}

// MountPublic registers the routes that do not require authentication.
func (h *Handler) MountPublic(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/accounts", h.register)
	mux.HandleFunc("POST /api/sessions", h.login)
}

// MountAuthed registers the routes that require a valid session. Each is wrapped
// in RequireAuth, and state-changing routes also in RequireCSRF.
func (h *Handler) MountAuthed(mux *http.ServeMux) {
	mux.Handle("DELETE /api/sessions/current", h.protect(h.logout))
	mux.Handle("GET /api/account", h.protectRead(h.getAccount))
	mux.Handle("PUT /api/account/timezone", h.protect(h.putTimezone))
	mux.Handle("PUT /api/account/password", h.protect(h.putPassword))
}

// protect wraps fn in auth + CSRF (for unsafe methods).
func (h *Handler) protect(fn http.HandlerFunc) http.Handler {
	return h.RequireAuth(h.RequireCSRF(fn))
}

// protectRead wraps fn in auth only (safe methods).
func (h *Handler) protectRead(fn http.HandlerFunc) http.Handler {
	return h.RequireAuth(fn)
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Timezone string `json:"timezone"`
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	if !h.throttleByIP(w, r) {
		return
	}
	var req registerRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	profile, sess, err := h.svc.Register(r.Context(), RegisterInput(req))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	if err := h.startSession(w, sess); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, profileBody(profile))
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	if !h.throttleByIP(w, r) {
		return
	}
	var req loginRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	fields := map[string]string{}
	if strings.TrimSpace(req.Email) == "" {
		fields["email"] = "email is required"
	}
	if req.Password == "" {
		fields["password"] = "password is required"
	}
	if len(fields) > 0 {
		httpx.WriteError(w, r, httpx.ValidationError(fields))
		return
	}

	key := strings.ToLower(strings.TrimSpace(req.Email)) + "|" + clientIP(r)
	if !h.loginLimiter.Allowed(key) {
		httpx.WriteError(w, r, httpx.NewError(http.StatusTooManyRequests, httpx.CodeRateLimited,
			"Too many attempts. Try again later."))
		return
	}

	sess, err := h.svc.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			h.loginLimiter.Fail(key)
		}
		writeServiceError(w, r, err)
		return
	}

	h.loginLimiter.Reset(key)
	if err := h.startSession(w, sess); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// logout ends the current session and clears the auth and CSRF cookies.
func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(SessionCookieName); err == nil {
		if err := h.svc.EndSession(r.Context(), cookie.Value); err != nil {
			httpx.WriteError(w, r, err)
			return
		}
	}
	h.clearSessionCookie(w)
	h.clearCSRFCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// getAccount returns the caller's own profile.
func (h *Handler) getAccount(w http.ResponseWriter, r *http.Request) {
	id, _ := reqctx.IdentityFrom(r.Context())
	profile, err := h.svc.Read(r.Context(), id.AccountID)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, profileBody(profile))
}

type timezoneRequest struct {
	Timezone string `json:"timezone"`
}

func (h *Handler) putTimezone(w http.ResponseWriter, r *http.Request) {
	var req timezoneRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	id, _ := reqctx.IdentityFrom(r.Context())
	if err := h.svc.SetTimezone(r.Context(), id.AccountID, req.Timezone); err != nil {
		writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type passwordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *Handler) putPassword(w http.ResponseWriter, r *http.Request) {
	var req passwordRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	id, _ := reqctx.IdentityFrom(r.Context())
	if err := h.svc.ChangePassword(r.Context(), id.AccountID, req.CurrentPassword, req.NewPassword); err != nil {
		writeServiceError(w, r, err)
		return
	}
	// Every session, including this one, is now gone (ADR-0004). Clear cookies.
	h.clearSessionCookie(w)
	h.clearCSRFCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// startSession sets the session cookie and a fresh double-submit CSRF cookie.
func (h *Handler) startSession(w http.ResponseWriter, s Session) error {
	h.setSessionCookie(w, s)
	token, err := newToken()
	if err != nil {
		return err
	}
	h.setCSRFCookie(w, token)
	return nil
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, s Session) {
	//nolint:gosec // G124: Secure is env-conditional (false only in local dev); HttpOnly + SameSite are set
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    s.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  s.ExpiresAt,
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	//nolint:gosec // G124: see setSessionCookie
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func profileBody(p Profile) map[string]string {
	return map[string]string{"email": p.Email, "timezone": p.Timezone}
}

// writeServiceError maps a Service error to the HTTP error envelope.
func writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var verr *ValidationError
	switch {
	case errors.As(err, &verr):
		httpx.WriteError(w, r, httpx.ValidationError(verr.Fields))
	case errors.Is(err, ErrEmailTaken):
		httpx.WriteError(w, r, httpx.NewError(http.StatusConflict, httpx.CodeEmailTaken,
			"An account with this email already exists"))
	case errors.Is(err, ErrInvalidCredentials):
		httpx.WriteError(w, r, httpx.NewError(http.StatusUnauthorized, httpx.CodeInvalidCredentials,
			"Invalid email or password"))
	case errors.Is(err, ErrSessionInvalid), errors.Is(err, ErrAccountNotFound):
		httpx.WriteError(w, r, httpx.NewError(http.StatusUnauthorized, httpx.CodeUnauthenticated,
			"Authentication required"))
	default:
		httpx.WriteError(w, r, err) // generic 500, logged, scrubbed
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
