// Package categories owns the categories table — a flat, per-account set of
// user-defined labels, each with an optional colour and icon. A category groups
// time blocks, tasks, habits and goals (and, later, notes and events); every other
// domain module reaches it only through Service (ADR-0002, ADR-0009).
package categories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Category is one label. Colour and Icon are a key from a fixed client-side set
// (empty when unset) — the backend stores and returns them but never interprets
// them (v1.md §2, amended by ADR-0009).
type Category struct {
	ID         uuid.UUID
	Name       string
	Colour     string
	Icon       string
	ArchivedAt *time.Time // nil while active
}

// Input is the data to create a category. Colour and Icon are optional; an empty
// value leaves them unset.
type Input struct {
	Name   string
	Colour string
	Icon   string
}

// UpdateInput is a partial update: a nil field is left unchanged; a non-nil field
// (including a pointer to "") sets that field to the given value — an explicit ""
// clears Colour/Icon, but Name must still be non-blank after trimming (R3,
// docs/left.md's Category C.1 gap: a rename-only PATCH must not wipe colour/icon).
type UpdateInput struct {
	Name   *string
	Colour *string
	Icon   *string
}

// Sentinel errors mapped to HTTP status codes by the handler.
var (
	ErrNameTaken = errors.New("categories: an active category with this name already exists")
	ErrNotFound  = errors.New("categories: category not found")
)

// ValidationError carries per-field messages for a 400 VALIDATION_ERROR.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "categories: validation failed" }

// Service is the categories module's published interface. Every method is scoped to
// the account id the caller passes, which callers take only from the request
// context (ADR-0004).
type Service interface {
	// Create adds an active category. ErrNameTaken if an active one already has
	// that name (case-insensitive). *ValidationError for a bad field.
	Create(ctx context.Context, accountID uuid.UUID, in Input) (Category, error)

	// Update partially updates an active category — a nil field in in is left
	// unchanged (R3). ErrNotFound if it is missing, not the caller's, or archived;
	// ErrNameTaken on a name collision; *ValidationError for a bad field.
	Update(ctx context.Context, accountID, categoryID uuid.UUID, in UpdateInput) (Category, error)

	// Archive archives an active category so it is no longer offered. Items already
	// assigned to it keep it (v1.md §2). ErrNotFound if it is missing, not the
	// caller's, or already archived.
	Archive(ctx context.Context, accountID, categoryID uuid.UUID) error

	// List returns the caller's active categories, name-ordered.
	List(ctx context.Context, accountID uuid.UUID) ([]Category, error)

	// ListAll returns every category the caller owns, active and archived,
	// name-ordered (M8 export).
	ListAll(ctx context.Context, accountID uuid.UUID) ([]Category, error)

	// AssignableToAccount reports whether categoryID names a category that the
	// account owns and that is still active — used by other modules to validate a
	// supplied category_id (ADR-0009 CategoryChecker).
	AssignableToAccount(ctx context.Context, accountID, categoryID uuid.UUID) (bool, error)

	// NamesForAccount returns every category the account owns (active and archived)
	// keyed by id — used by read models to label items whose category may since
	// have been archived.
	NamesForAccount(ctx context.Context, accountID uuid.UUID) (map[uuid.UUID]string, error)
}

// Counter is implemented by any module that owns account-scoped items which may
// reference a category — tasks, habits, goals, timeline, and later notes/events.
// cmd/server composes GET /api/categories/overview from one Counter per module, so
// categories itself never needs to know another module's schema (ADR-0009).
type Counter interface {
	// CountByCategory returns, for every category the account has assigned at
	// least one item to, how many of the caller's items reference it. A category
	// with no items is simply absent from the map.
	CountByCategory(ctx context.Context, accountID uuid.UUID) (map[uuid.UUID]int, error)
}
