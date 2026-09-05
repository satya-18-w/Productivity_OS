package notes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/satya-18-w/productivity-os/internal/notes/notesdb"
)

const (
	maxTitleLen = 200
	maxBodyLen  = 20000
)

type service struct {
	q *notesdb.Queries
}

// NewService builds the notes service over a connection pool.
func NewService(pool *pgxpool.Pool) Service {
	return &service{q: notesdb.New(pool)}
}

func validateInput(in NoteInput) (NoteInput, *ValidationError) {
	fields := map[string]string{}
	title := strings.TrimSpace(in.Title)
	switch {
	case title == "":
		fields["title"] = "title is required"
	case len(title) > maxTitleLen:
		fields["title"] = "title must be at most 200 characters"
	}
	if len(in.Body) > maxBodyLen {
		fields["body"] = "body must be at most 20000 characters"
	}
	if len(fields) > 0 {
		return NoteInput{}, &ValidationError{Fields: fields}
	}
	return NoteInput{Title: title, Body: in.Body}, nil
}

func (s *service) CreateNote(ctx context.Context, accountID uuid.UUID, raw NoteInput) (Note, error) {
	in, verr := validateInput(raw)
	if verr != nil {
		return Note{}, verr
	}
	row, err := s.q.CreateNote(ctx, notesdb.CreateNoteParams{
		AccountID: accountID,
		Title:     in.Title,
		Body:      in.Body,
	})
	if err != nil {
		return Note{}, fmt.Errorf("create note: %w", err)
	}
	return toNote(row.ID, row.Title, row.Body, row.CreatedAt, row.UpdatedAt), nil
}

func (s *service) UpdateNote(ctx context.Context, accountID, noteID uuid.UUID, raw NoteInput) error {
	in, verr := validateInput(raw)
	if verr != nil {
		return verr
	}
	rows, err := s.q.UpdateNoteFields(ctx, notesdb.UpdateNoteFieldsParams{
		AccountID: accountID,
		ID:        noteID,
		Title:     in.Title,
		Body:      in.Body,
	})
	if err != nil {
		return fmt.Errorf("update note: %w", err)
	}
	if rows == 0 {
		return ErrNoteNotFound
	}
	return nil
}

func (s *service) DeleteNote(ctx context.Context, accountID, noteID uuid.UUID) error {
	rows, err := s.q.DeleteNote(ctx, notesdb.DeleteNoteParams{AccountID: accountID, ID: noteID})
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	if rows == 0 {
		return ErrNoteNotFound
	}
	return nil
}

func (s *service) ListNotes(ctx context.Context, accountID uuid.UUID) ([]Note, error) {
	rows, err := s.q.ListNotes(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	out := make([]Note, len(rows))
	for i, r := range rows {
		out[i] = toNote(r.ID, r.Title, r.Body, r.CreatedAt, r.UpdatedAt)
	}
	return out, nil
}

func toNote(id uuid.UUID, title, body string, created, updated pgtype.Timestamptz) Note {
	return Note{
		ID:        id,
		Title:     title,
		Body:      body,
		CreatedAt: created.Time.UTC().Format(time.RFC3339),
		UpdatedAt: updated.Time.UTC().Format(time.RFC3339),
	}
}
