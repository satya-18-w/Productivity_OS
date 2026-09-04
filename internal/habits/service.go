package habits

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/satya-18-w/productivity-os/internal/habits/habitsdb"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

const (
	maxNameLen   = 100
	streakWindow = 400 // days of history loaded for the streak / 30-day count
)

type service struct {
	pool *pgxpool.Pool
	q    *habitsdb.Queries
	zone AccountZone
}

// NewService builds the habits service over a connection pool.
func NewService(pool *pgxpool.Pool, zone AccountZone) Service {
	return &service{pool: pool, q: habitsdb.New(pool), zone: zone}
}

func validateName(raw string) (string, *ValidationError) {
	n := strings.TrimSpace(raw)
	switch {
	case n == "":
		return "", &ValidationError{Fields: map[string]string{"name": "name is required"}}
	case len(n) > maxNameLen:
		return "", &ValidationError{Fields: map[string]string{"name": "name must be at most 100 characters"}}
	default:
		return n, nil
	}
}

func (s *service) CreateHabit(ctx context.Context, accountID uuid.UUID, name string) (Habit, error) {
	n, verr := validateName(name)
	if verr != nil {
		return Habit{}, verr
	}
	row, err := s.q.CreateHabit(ctx, habitsdb.CreateHabitParams{AccountID: accountID, Name: n})
	if err != nil {
		return Habit{}, fmt.Errorf("create habit: %w", err)
	}
	return Habit{ID: row.ID, Name: row.Name, ArchivedAt: nullTime(row.ArchivedAt)}, nil
}

func (s *service) ArchiveHabit(ctx context.Context, accountID, habitID uuid.UUID) error {
	return s.setArchived(ctx, accountID, habitID, pgtype.Timestamptz{Time: time.Now(), Valid: true})
}

func (s *service) UnarchiveHabit(ctx context.Context, accountID, habitID uuid.UUID) error {
	return s.setArchived(ctx, accountID, habitID, pgtype.Timestamptz{})
}

func (s *service) setArchived(ctx context.Context, accountID, habitID uuid.UUID, at pgtype.Timestamptz) error {
	rows, err := s.q.SetHabitArchived(ctx, habitsdb.SetHabitArchivedParams{
		AccountID: accountID, ID: habitID, ArchivedAt: at,
	})
	if err != nil {
		return fmt.Errorf("set archived: %w", err)
	}
	if rows == 0 {
		return ErrHabitNotFound
	}
	return nil
}

func (s *service) MarkComplete(ctx context.Context, accountID, habitID uuid.UUID, date timezone.Date) error {
	if err := s.assertOwned(ctx, accountID, habitID); err != nil {
		return err
	}
	if err := s.q.MarkCompletion(ctx, habitsdb.MarkCompletionParams{
		HabitID: habitID, AccountID: accountID, OnDate: pgDate(date),
	}); err != nil {
		return fmt.Errorf("mark completion: %w", err)
	}
	return nil
}

func (s *service) UnmarkComplete(ctx context.Context, accountID, habitID uuid.UUID, date timezone.Date) error {
	if err := s.assertOwned(ctx, accountID, habitID); err != nil {
		return err
	}
	if err := s.q.UnmarkCompletion(ctx, habitsdb.UnmarkCompletionParams{
		HabitID: habitID, OnDate: pgDate(date),
	}); err != nil {
		return fmt.Errorf("unmark completion: %w", err)
	}
	return nil
}

func (s *service) assertOwned(ctx context.Context, accountID, habitID uuid.UUID) error {
	n, err := s.q.HabitBelongsToAccount(ctx, habitsdb.HabitBelongsToAccountParams{
		AccountID: accountID, ID: habitID,
	})
	if err != nil {
		return fmt.Errorf("check habit: %w", err)
	}
	if n == 0 {
		return ErrHabitNotFound
	}
	return nil
}

func (s *service) ListActive(ctx context.Context, accountID uuid.UUID, viewDate timezone.Date) ([]HabitView, error) {
	loc, err := s.zone.Zone(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("resolve account zone: %w", err)
	}
	today := timezone.Today(loc)
	since := today.AddDays(-streakWindow)
	last30From := today.AddDays(-29)

	rows, err := s.q.ListActiveHabits(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("list habits: %w", err)
	}

	out := make([]HabitView, 0, len(rows))
	for _, h := range rows {
		dates, err := s.q.ListCompletionDatesSince(ctx, habitsdb.ListCompletionDatesSinceParams{
			HabitID: h.ID, Since: pgDate(since),
		})
		if err != nil {
			return nil, fmt.Errorf("list completions: %w", err)
		}

		set := make(map[string]struct{}, len(dates))
		for _, d := range dates {
			set[timezone.Date{Year: d.Time.Year(), Month: d.Time.Month(), Day: d.Time.Day()}.String()] = struct{}{}
		}

		_, done := set[viewDate.String()]
		out = append(out, HabitView{
			ID:              h.ID,
			Name:            h.Name,
			CurrentStreak:   currentStreak(set, today),
			CompletedOnDate: done,
			Last30Days:      countInRange(set, last30From, today),
		})
	}
	return out, nil
}

func (s *service) ListArchived(ctx context.Context, accountID uuid.UUID) ([]Habit, error) {
	rows, err := s.q.ListArchivedHabits(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("list archived habits: %w", err)
	}
	out := make([]Habit, len(rows))
	for i, h := range rows {
		out[i] = Habit{ID: h.ID, Name: h.Name, ArchivedAt: nullTime(h.ArchivedAt)}
	}
	return out, nil
}

func pgDate(d timezone.Date) pgtype.Date {
	return pgtype.Date{
		Time:  time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}
}

func nullTime(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}
