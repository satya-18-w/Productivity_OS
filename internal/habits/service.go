package habits

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/satya-18-w/productivity-os/internal/habits/habitsdb"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

const (
	maxNameLen          = 100
	streakWindow        = 400 // days of history loaded for the streak / 30-day count
	maxHistoryRangeDays = 92  // R2, docs/left.md Phase 6
)

type service struct {
	pool *pgxpool.Pool
	q    *habitsdb.Queries
	zone AccountZone
	cats CategoryChecker
}

// NewService builds the habits service over a connection pool. cats validates a
// category_id a caller assigns (wired to categories.Service by cmd/server —
// ADR-0009).
func NewService(pool *pgxpool.Pool, zone AccountZone, cats CategoryChecker) Service {
	return &service{pool: pool, q: habitsdb.New(pool), zone: zone, cats: cats}
}

func (s *service) assertAssignableCategory(ctx context.Context, accountID uuid.UUID, categoryID *uuid.UUID) error {
	if categoryID == nil {
		return nil
	}
	ok, err := s.cats.AssignableToAccount(ctx, accountID, *categoryID)
	if err != nil {
		return fmt.Errorf("check category: %w", err)
	}
	if !ok {
		return &ValidationError{Fields: map[string]string{"category_id": "category not found or archived"}}
	}
	return nil
}

const maxTargetLen = 100

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

// validateTarget trims and bounds an optional target descriptor; a nil or
// empty-after-trim target is left unset (nil), not stored as "".
func validateTarget(raw *string) (*string, *ValidationError) {
	if raw == nil {
		return nil, nil
	}
	t := strings.TrimSpace(*raw)
	if t == "" {
		return nil, nil
	}
	if len(t) > maxTargetLen {
		return nil, &ValidationError{Fields: map[string]string{"target": "target must be at most 100 characters"}}
	}
	return &t, nil
}

func (s *service) CreateHabit(ctx context.Context, accountID uuid.UUID, in HabitInput) (Habit, error) {
	n, verr := validateName(in.Name)
	if verr != nil {
		return Habit{}, verr
	}
	target, verr := validateTarget(in.Target)
	if verr != nil {
		return Habit{}, verr
	}
	if err := s.assertAssignableCategory(ctx, accountID, in.CategoryID); err != nil {
		return Habit{}, err
	}
	row, err := s.q.CreateHabit(ctx, habitsdb.CreateHabitParams{
		AccountID: accountID, Name: n, CategoryID: toPgUUID(in.CategoryID), Target: toPgText(target),
	})
	if err != nil {
		return Habit{}, fmt.Errorf("create habit: %w", err)
	}
	return Habit{
		ID: row.ID, Name: row.Name, CategoryID: fromPgUUID(row.CategoryID), Target: fromPgText(row.Target),
		ArchivedAt: nullTime(row.ArchivedAt), CreatedAt: row.CreatedAt.Time,
	}, nil
}

func (s *service) UpdateHabit(ctx context.Context, accountID, habitID uuid.UUID, name string, rawTarget *string) (Habit, error) {
	n, verr := validateName(name)
	if verr != nil {
		return Habit{}, verr
	}
	target, verr := validateTarget(rawTarget)
	if verr != nil {
		return Habit{}, verr
	}
	row, err := s.q.UpdateHabitFields(ctx, habitsdb.UpdateHabitFieldsParams{
		AccountID: accountID, ID: habitID, Name: n, Target: toPgText(target),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Habit{}, ErrHabitNotFound
		}
		return Habit{}, fmt.Errorf("update habit: %w", err)
	}
	return Habit{
		ID: row.ID, Name: row.Name, CategoryID: fromPgUUID(row.CategoryID), Target: fromPgText(row.Target),
		ArchivedAt: nullTime(row.ArchivedAt), CreatedAt: row.CreatedAt.Time,
	}, nil
}

func (s *service) SetHabitCategory(ctx context.Context, accountID, habitID uuid.UUID, categoryID *uuid.UUID) error {
	if err := s.assertOwned(ctx, accountID, habitID); err != nil {
		return err
	}
	if err := s.assertAssignableCategory(ctx, accountID, categoryID); err != nil {
		return err
	}
	rows, err := s.q.SetHabitCategory(ctx, habitsdb.SetHabitCategoryParams{
		AccountID: accountID, ID: habitID, CategoryID: toPgUUID(categoryID),
	})
	if err != nil {
		return fmt.Errorf("set habit category: %w", err)
	}
	if rows == 0 {
		return ErrHabitNotFound
	}
	return nil
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
			CategoryID:      fromPgUUID(h.CategoryID),
			Target:          fromPgText(h.Target),
			CurrentStreak:   currentStreak(set, today),
			CompletedOnDate: done,
			Last30Days:      countInRange(set, last30From, today),
		})
	}
	return out, nil
}

// Week batches the "This Week" grid into one query per habit instead of one
// ListActive-shaped call per day (docs/left.md Phase 6 — mirrors ListActive's own
// streak computation exactly, just scoped to the requested ISO week's 7 dates for
// the Completed field).
func (s *service) Week(ctx context.Context, accountID uuid.UUID, date timezone.Date) (WeekView, error) {
	loc, err := s.zone.Zone(ctx, accountID)
	if err != nil {
		return WeekView{}, fmt.Errorf("resolve account zone: %w", err)
	}
	start, _ := timezone.ISOWeekWindow(date, loc)
	weekStart := timezone.DateAt(start, loc)
	days := make([]timezone.Date, 7)
	for i := range days {
		days[i] = weekStart.AddDays(i)
	}

	today := timezone.Today(loc)
	since := today.AddDays(-streakWindow)

	rows, err := s.q.ListActiveHabits(ctx, accountID)
	if err != nil {
		return WeekView{}, fmt.Errorf("list habits: %w", err)
	}

	habitsOut := make([]HabitWeekEntry, 0, len(rows))
	for _, h := range rows {
		dates, err := s.q.ListCompletionDatesSince(ctx, habitsdb.ListCompletionDatesSinceParams{
			HabitID: h.ID, Since: pgDate(since),
		})
		if err != nil {
			return WeekView{}, fmt.Errorf("list completions: %w", err)
		}
		set := make(map[string]struct{}, len(dates))
		for _, d := range dates {
			set[timezone.Date{Year: d.Time.Year(), Month: d.Time.Month(), Day: d.Time.Day()}.String()] = struct{}{}
		}

		completed := make([]timezone.Date, 0, len(days))
		for _, d := range days {
			if _, ok := set[d.String()]; ok {
				completed = append(completed, d)
			}
		}
		habitsOut = append(habitsOut, HabitWeekEntry{
			HabitID: h.ID, Name: h.Name, CurrentStreak: currentStreak(set, today), Completed: completed,
		})
	}

	archivedRows, err := s.q.ListArchivedHabits(ctx, accountID)
	if err != nil {
		return WeekView{}, fmt.Errorf("list archived habits: %w", err)
	}
	archived := make([]ArchivedHabitName, len(archivedRows))
	for i, h := range archivedRows {
		archived[i] = ArchivedHabitName{HabitID: h.ID, Name: h.Name}
	}

	return WeekView{WeekStart: weekStart, Days: days, Habits: habitsOut, Archived: archived}, nil
}

func (s *service) ListArchived(ctx context.Context, accountID uuid.UUID) ([]Habit, error) {
	rows, err := s.q.ListArchivedHabits(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("list archived habits: %w", err)
	}
	out := make([]Habit, len(rows))
	for i, h := range rows {
		out[i] = Habit{
			ID: h.ID, Name: h.Name, CategoryID: fromPgUUID(h.CategoryID), Target: fromPgText(h.Target),
			ArchivedAt: nullTime(h.ArchivedAt), CreatedAt: h.CreatedAt.Time,
		}
	}
	return out, nil
}

func (s *service) CompletionCountsInRange(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) ([]RangeCount, error) {
	if to.Before(from) {
		return nil, &ValidationError{Fields: map[string]string{"to": "must not be before from"}}
	}
	rows, err := s.q.CompletionCountsInRange(ctx, habitsdb.CompletionCountsInRangeParams{
		AccountID: accountID, FromDate: pgDate(from), ToDate: pgDate(to),
	})
	if err != nil {
		return nil, fmt.Errorf("completion counts in range: %w", err)
	}
	out := make([]RangeCount, len(rows))
	for i, r := range rows {
		out[i] = RangeCount{HabitID: r.HabitID, Name: r.Name, Count: int(r.Total)}
	}
	return out, nil
}

func (s *service) ListAll(ctx context.Context, accountID uuid.UUID) ([]Habit, error) {
	rows, err := s.q.ListAllHabits(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("list all habits: %w", err)
	}
	out := make([]Habit, len(rows))
	for i, h := range rows {
		out[i] = Habit{
			ID: h.ID, Name: h.Name, CategoryID: fromPgUUID(h.CategoryID), Target: fromPgText(h.Target),
			ArchivedAt: nullTime(h.ArchivedAt), CreatedAt: h.CreatedAt.Time,
		}
	}
	return out, nil
}

func (s *service) AllCompletions(ctx context.Context, accountID uuid.UUID) ([]HabitCompletion, error) {
	rows, err := s.q.ListAllCompletions(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("list all completions: %w", err)
	}
	out := make([]HabitCompletion, len(rows))
	for i, r := range rows {
		out[i] = HabitCompletion{
			HabitID: r.HabitID,
			Date:    timezone.Date{Year: r.OnDate.Time.Year(), Month: r.OnDate.Time.Month(), Day: r.OnDate.Time.Day()},
		}
	}
	return out, nil
}

func (s *service) History(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) ([]HabitHistoryEntry, error) {
	if verr := validateHistoryRange(from, to); verr != nil {
		return nil, verr
	}
	rows, err := s.q.HabitHistory(ctx, habitsdb.HabitHistoryParams{
		AccountID: accountID, FromDate: pgDate(from), ToDate: pgDate(to),
	})
	if err != nil {
		return nil, fmt.Errorf("habit history: %w", err)
	}

	order := make([]uuid.UUID, 0)
	byID := make(map[uuid.UUID]*HabitHistoryEntry)
	for _, r := range rows {
		e, ok := byID[r.HabitID]
		if !ok {
			e = &HabitHistoryEntry{HabitID: r.HabitID, Name: r.Name, Archived: r.ArchivedAt.Valid, Completions: []timezone.Date{}}
			byID[r.HabitID] = e
			order = append(order, r.HabitID)
		}
		if r.OnDate.Valid {
			e.Completions = append(e.Completions, timezone.Date{Year: r.OnDate.Time.Year(), Month: r.OnDate.Time.Month(), Day: r.OnDate.Time.Day()})
		}
	}

	out := make([]HabitHistoryEntry, len(order))
	for i, id := range order {
		out[i] = *byID[id]
	}
	return out, nil
}

func validateHistoryRange(from, to timezone.Date) *ValidationError {
	if to.Before(from) {
		return &ValidationError{Fields: map[string]string{"to": "must not be before from"}}
	}
	if rangeDayCount(from, to) > maxHistoryRangeDays {
		return &ValidationError{Fields: map[string]string{
			"range": fmt.Sprintf("must not exceed %d days", maxHistoryRangeDays),
		}}
	}
	return nil
}

// rangeDayCount is the inclusive day count of [from, to] — callers only invoke it
// after confirming to is not before from.
func rangeDayCount(from, to timezone.Date) int {
	n := 1
	for d := from; d.Before(to); d = d.Next() {
		n++
	}
	return n
}

func (s *service) CountByCategory(ctx context.Context, accountID uuid.UUID) (map[uuid.UUID]int, error) {
	rows, err := s.q.CountHabitsByCategory(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("count habits by category: %w", err)
	}
	out := make(map[uuid.UUID]int, len(rows))
	for _, r := range rows {
		if id := fromPgUUID(r.CategoryID); id != nil {
			out[*id] = int(r.Total)
		}
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

func toPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func fromPgText(v pgtype.Text) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}
