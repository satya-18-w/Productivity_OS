// Package reqctx carries the authenticated principal through the request context.
// The auth middleware (in internal/account) sets it; every domain module's
// handlers read it. The acting account id comes from here and nowhere else
// (ADR-0004, v1.md N3).
package reqctx

import (
	"context"

	"github.com/google/uuid"
)

// Identity is the authenticated principal.
type Identity struct {
	AccountID uuid.UUID
}

type ctxKey int

const identityKey ctxKey = iota

// WithIdentity returns a context carrying id.
func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, identityKey, id)
}

// IdentityFrom returns the identity in ctx, and false if none is set.
func IdentityFrom(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(identityKey).(Identity)
	return id, ok
}
