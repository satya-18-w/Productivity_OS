// Package notes owns free-form notes (v1.md §15). Service is its entire public
// surface (ADR-0002). A note is plain text only and linked to nothing else.
package notes

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// Note is a single note.
type Note struct {
	ID        uuid.UUID
	Title     string
	Body      string
	CreatedAt string // RFC3339
	UpdatedAt string // RFC3339
}

// NoteInput is the editable field set.
type NoteInput struct {
	Title string
	Body  string
}

// ErrNoteNotFound is returned when a note is missing or not the caller's.
var ErrNoteNotFound = errors.New("notes: note not found")

// ValidationError carries per-field messages for a 400 VALIDATION_ERROR.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "notes: validation failed" }

// Service is the notes module's published interface.
type Service interface {
	// CreateNote adds a note.
	CreateNote(ctx context.Context, accountID uuid.UUID, in NoteInput) (Note, error)

	// UpdateNote replaces a note's title and body. ErrNoteNotFound if it is
	// missing or not the caller's.
	UpdateNote(ctx context.Context, accountID, noteID uuid.UUID, in NoteInput) error

	// DeleteNote removes a note permanently — there is no trash/soft-delete.
	DeleteNote(ctx context.Context, accountID, noteID uuid.UUID) error

	// ListNotes returns the caller's notes, newest first. Also implements
	// export.NotesReader (MX4) — there is no active/archived split to require a
	// separate export-only listing, mirroring goals.Service.ListGoals.
	ListNotes(ctx context.Context, accountID uuid.UUID) ([]Note, error)
}
