// Package account owns the account and session concepts and all authentication
// logic. Service is its entire public surface (ADR-0002); nothing else in the
// package is imported by other modules.
package account

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
)

// Profile is an account's own visible data (v1.md §1: email and timezone only).
type Profile struct {
	Email    string
	Timezone string
}

// Session is a freshly issued session and its opaque token.
type Session struct {
	Token     string
	ExpiresAt time.Time
}

// RegisterInput is the data required to create an account.
type RegisterInput struct {
	Email    string
	Password string
	Timezone string // "" means "apply the default" (Q4)
}

// Sentinel errors. The HTTP layer maps these to status codes; it never forwards
// their text to clients for 5xx-adjacent cases.
var (
	ErrEmailTaken         = errors.New("account: email already registered")
	ErrInvalidCredentials = errors.New("account: invalid credentials")
	ErrSessionInvalid     = errors.New("account: session invalid or expired")
	ErrAccountNotFound    = errors.New("account: not found")
)

// ValidationError carries per-field messages for a 400 VALIDATION_ERROR.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "account: validation failed" }

// Service is the account module's published interface.
type Service interface {
	// Register creates an account and opens a session for it (registration logs
	// the client in). ErrEmailTaken if the email is in use; *ValidationError for
	// bad input.
	Register(ctx context.Context, in RegisterInput) (Profile, Session, error)

	// Authenticate verifies a credential pair and opens a session.
	// ErrInvalidCredentials for both an unknown email and a wrong password.
	Authenticate(ctx context.Context, email, plaintext string) (Session, error)

	// ResolveSession maps a session token to its account. ErrSessionInvalid if
	// the token is unknown or the session has expired.
	ResolveSession(ctx context.Context, token string) (reqctx.Identity, error)

	// EndSession deletes one session (logout). Deleting an unknown token is a
	// no-op.
	EndSession(ctx context.Context, token string) error

	// EndAllSessions deletes every session for an account.
	EndAllSessions(ctx context.Context, accountID uuid.UUID) error

	// ChangePassword verifies the current password, stores a new hash, and ends
	// every session for the account. ErrInvalidCredentials if current is wrong;
	// *ValidationError if next fails the policy.
	ChangePassword(ctx context.Context, accountID uuid.UUID, current, next string) error

	// SetTimezone stores a new IANA timezone. *ValidationError if invalid.
	SetTimezone(ctx context.Context, accountID uuid.UUID, tz string) error

	// Read returns the account's own profile. ErrAccountNotFound if it is gone.
	Read(ctx context.Context, accountID uuid.UUID) (Profile, error)
}
