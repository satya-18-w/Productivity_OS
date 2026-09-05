package categories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/satya-18-w/productivity-os/internal/categories/categoriesdb"
)

const (
	uniqueViolation = "23505"
	maxNameLen      = 60
	maxKeyLen       = 40 // colour / icon key
)

type service struct {
	q *categoriesdb.Queries
}

// NewService builds the categories service over a connection pool.
func NewService(pool *pgxpool.Pool) Service {
	return &service{q: categoriesdb.New(pool)}
}

func validate(in Input) (Input, *ValidationError) {
	fields := map[string]string{}

	name := strings.TrimSpace(in.Name)
	switch {
	case name == "":
		fields["name"] = "name is required"
	case len(name) > maxNameLen:
		fields["name"] = "name must be at most 60 characters"
	}

	colour := strings.TrimSpace(in.Colour)
	if len(colour) > maxKeyLen {
		fields["colour"] = "colour must be at most 40 characters"
	}

	icon := strings.TrimSpace(in.Icon)
	if len(icon) > maxKeyLen {
		fields["icon"] = "icon must be at most 40 characters"
	}

	if len(fields) > 0 {
		return Input{}, &ValidationError{Fields: fields}
	}
	return Input{Name: name, Colour: colour, Icon: icon}, nil
}

func text(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func (s *service) Create(ctx context.Context, accountID uuid.UUID, in Input) (Category, error) {
	clean, verr := validate(in)
	if verr != nil {
		return Category{}, verr
	}
	row, err := s.q.CreateCategory(ctx, categoriesdb.CreateCategoryParams{
		AccountID: accountID,
		Name:      clean.Name,
		Colour:    text(clean.Colour),
		Icon:      text(clean.Icon),
	})
	if err != nil {
		if isUnique(err) {
			return Category{}, ErrNameTaken
		}
		return Category{}, fmt.Errorf("create category: %w", err)
	}
	return Category{
		ID:         row.ID,
		Name:       row.Name,
		Colour:     row.Colour,
		Icon:       row.Icon,
		ArchivedAt: nullTime(row.ArchivedAt),
	}, nil
}

func (s *service) Update(ctx context.Context, accountID, categoryID uuid.UUID, in UpdateInput) (Category, error) {
	current, err := s.q.GetActiveCategory(ctx, categoriesdb.GetActiveCategoryParams{
		AccountID: accountID, ID: categoryID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Category{}, ErrNotFound
		}
		return Category{}, fmt.Errorf("get category: %w", err)
	}

	// Merge: a nil field leaves the current value untouched (R3).
	merged := Input{Name: current.Name, Colour: current.Colour, Icon: current.Icon}
	if in.Name != nil {
		merged.Name = *in.Name
	}
	if in.Colour != nil {
		merged.Colour = *in.Colour
	}
	if in.Icon != nil {
		merged.Icon = *in.Icon
	}
	clean, verr := validate(merged)
	if verr != nil {
		return Category{}, verr
	}

	rows, err := s.q.UpdateCategory(ctx, categoriesdb.UpdateCategoryParams{
		AccountID: accountID,
		ID:        categoryID,
		Name:      clean.Name,
		Colour:    text(clean.Colour),
		Icon:      text(clean.Icon),
	})
	if err != nil {
		if isUnique(err) {
			return Category{}, ErrNameTaken
		}
		return Category{}, fmt.Errorf("update category: %w", err)
	}
	if rows == 0 {
		return Category{}, ErrNotFound
	}
	return Category{ID: categoryID, Name: clean.Name, Colour: clean.Colour, Icon: clean.Icon}, nil
}

func (s *service) Archive(ctx context.Context, accountID, categoryID uuid.UUID) error {
	rows, err := s.q.ArchiveCategory(ctx, categoriesdb.ArchiveCategoryParams{
		AccountID: accountID, ID: categoryID,
	})
	if err != nil {
		return fmt.Errorf("archive category: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *service) List(ctx context.Context, accountID uuid.UUID) ([]Category, error) {
	rows, err := s.q.ListActiveCategories(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	out := make([]Category, len(rows))
	for i, r := range rows {
		out[i] = Category{
			ID:         r.ID,
			Name:       r.Name,
			Colour:     r.Colour,
			Icon:       r.Icon,
			ArchivedAt: nullTime(r.ArchivedAt),
		}
	}
	return out, nil
}

func (s *service) ListAll(ctx context.Context, accountID uuid.UUID) ([]Category, error) {
	rows, err := s.q.ListAllCategories(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("list all categories: %w", err)
	}
	out := make([]Category, len(rows))
	for i, r := range rows {
		out[i] = Category{
			ID:         r.ID,
			Name:       r.Name,
			Colour:     r.Colour,
			Icon:       r.Icon,
			ArchivedAt: nullTime(r.ArchivedAt),
		}
	}
	return out, nil
}

func (s *service) AssignableToAccount(ctx context.Context, accountID, categoryID uuid.UUID) (bool, error) {
	n, err := s.q.CountAssignableCategory(ctx, categoriesdb.CountAssignableCategoryParams{
		AccountID: accountID, ID: categoryID,
	})
	if err != nil {
		return false, fmt.Errorf("check category: %w", err)
	}
	return n > 0, nil
}

func (s *service) NamesForAccount(ctx context.Context, accountID uuid.UUID) (map[uuid.UUID]string, error) {
	rows, err := s.q.ListCategoryNames(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("list category names: %w", err)
	}
	out := make(map[uuid.UUID]string, len(rows))
	for _, r := range rows {
		out[r.ID] = r.Name
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
