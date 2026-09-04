package timeline

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/satya-18-w/productivity-os/internal/timeline/timelinedb"
)

const (
	uniqueViolation = "23505"
	maxNameLen      = 60
)

type service struct {
	pool *pgxpool.Pool
	q    *timelinedb.Queries
	zone AccountZone
}

// NewService builds the timeline service over a connection pool. zone resolves an
// account's timezone (wired to the account module by cmd/server).
func NewService(pool *pgxpool.Pool, zone AccountZone) Service {
	return &service{pool: pool, q: timelinedb.New(pool), zone: zone}
}

func validateName(raw string) (string, *ValidationError) {
	n := strings.TrimSpace(raw)
	switch {
	case n == "":
		return "", &ValidationError{Fields: map[string]string{"name": "name is required"}}
	case len(n) > maxNameLen:
		return "", &ValidationError{Fields: map[string]string{"name": "name must be at most 60 characters"}}
	default:
		return n, nil
	}
}

func (s *service) CreateCategory(ctx context.Context, accountID uuid.UUID, name string) (Category, error) {
	n, verr := validateName(name)
	if verr != nil {
		return Category{}, verr
	}
	row, err := s.q.CreateCategory(ctx, timelinedb.CreateCategoryParams{AccountID: accountID, Name: n})
	if err != nil {
		if isUnique(err) {
			return Category{}, ErrCategoryNameTaken
		}
		return Category{}, fmt.Errorf("create category: %w", err)
	}
	return Category{ID: row.ID, Name: row.Name, ArchivedAt: nullTime(row.ArchivedAt)}, nil
}

func (s *service) RenameCategory(ctx context.Context, accountID, categoryID uuid.UUID, name string) error {
	n, verr := validateName(name)
	if verr != nil {
		return verr
	}
	rows, err := s.q.RenameCategory(ctx, timelinedb.RenameCategoryParams{
		AccountID: accountID, ID: categoryID, Name: n,
	})
	if err != nil {
		if isUnique(err) {
			return ErrCategoryNameTaken
		}
		return fmt.Errorf("rename category: %w", err)
	}
	if rows == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func (s *service) ArchiveCategory(ctx context.Context, accountID, categoryID uuid.UUID) error {
	rows, err := s.q.ArchiveCategory(ctx, timelinedb.ArchiveCategoryParams{
		AccountID: accountID, ID: categoryID,
	})
	if err != nil {
		return fmt.Errorf("archive category: %w", err)
	}
	if rows == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func (s *service) ListActiveCategories(ctx context.Context, accountID uuid.UUID) ([]Category, error) {
	rows, err := s.q.ListActiveCategories(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	out := make([]Category, len(rows))
	for i, r := range rows {
		out[i] = Category{ID: r.ID, Name: r.Name, ArchivedAt: nullTime(r.ArchivedAt)}
	}
	return out, nil
}

func isUnique(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolation
}

func nullTime(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}

func toPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func fromPgUUID(v pgtype.UUID) *uuid.UUID {
	if !v.Valid {
		return nil
	}
	id := uuid.UUID(v.Bytes)
	return &id
}

func fromPgText(v pgtype.Text) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}
