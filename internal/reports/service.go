package reports

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/habits"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
	"github.com/satya-18-w/productivity-os/internal/timeline"
)

// maxReportRangeDays bounds the combined Report — an unbounded range over six
// composed queries is the one place in this module worth a server-side limit
// (docs/left.md Phase 9). The five sub-report methods stay unbounded, matching
// their existing M7 test coverage.
const maxReportRangeDays = 366

type service struct {
	timeline TimelineReader
	habits   HabitsReader
	tasks    TasksReader
	zone     AccountZone
}

// NewService builds the reports service over the other modules' published
// interfaces. It owns no connection pool — it never queries a table directly.
func NewService(timelineReader TimelineReader, habitsReader HabitsReader, tasksReader TasksReader, zone AccountZone) Service {
	return &service{timeline: timelineReader, habits: habitsReader, tasks: tasksReader, zone: zone}
}

func validateRange(from, to timezone.Date) *ValidationError {
	if to.Before(from) {
		return &ValidationError{Fields: map[string]string{"to": "must not be before from"}}
	}
	return nil
}

func (s *service) TimeByCategory(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) ([]CategoryTimeRow, error) {
	if verr := validateRange(from, to); verr != nil {
		return nil, verr
	}
	cmp, err := s.timeline.ComparisonRange(ctx, accountID, from, to)
	if err != nil {
		return nil, fmt.Errorf("comparison range: %w", err)
	}
	out := make([]CategoryTimeRow, len(cmp.Categories))
	for i, c := range cmp.Categories {
		out[i] = CategoryTimeRow{CategoryID: c.CategoryID, CategoryName: c.CategoryName, ActualSeconds: c.ActualSeconds}
	}
	return out, nil
}

func (s *service) PlannedVsActualByCategory(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) ([]timeline.CategoryTotals, error) {
	if verr := validateRange(from, to); verr != nil {
		return nil, verr
	}
	cmp, err := s.timeline.ComparisonRange(ctx, accountID, from, to)
	if err != nil {
		return nil, fmt.Errorf("comparison range: %w", err)
	}
	return cmp.Categories, nil
}

func (s *service) HabitCompletion(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) ([]HabitCompletionRow, error) {
	if verr := validateRange(from, to); verr != nil {
		return nil, verr
	}
	counts, err := s.habits.CompletionCountsInRange(ctx, accountID, from, to)
	if err != nil {
		return nil, fmt.Errorf("habit completion counts: %w", err)
	}
	all, err := s.habits.ListAll(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("list habits: %w", err)
	}
	loc, err := s.zone.Zone(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("resolve account zone: %w", err)
	}
	byID := make(map[uuid.UUID]habits.Habit, len(all))
	for _, h := range all {
		byID[h.ID] = h
	}

	out := make([]HabitCompletionRow, len(counts))
	for i, c := range counts {
		out[i] = HabitCompletionRow{
			HabitID: c.HabitID, Name: c.Name, CompletedDays: c.Count,
			RangeDays: activeDaysInRange(byID[c.HabitID], from, to, loc),
		}
	}
	return out, nil
}

// activeDaysInRange is the inclusive day count of [from, to] clipped to the days h
// actually existed (from its CreatedAt) and was active (up to the day before
// ArchivedAt, if archived) — never more than the full range, and 0 if the habit's
// active window and the queried range don't overlap at all.
func activeDaysInRange(h habits.Habit, from, to timezone.Date, loc *time.Location) int {
	start := from
	if created := timezone.DateAt(h.CreatedAt, loc); start.Before(created) {
		start = created
	}
	end := to
	if h.ArchivedAt != nil {
		if lastActive := timezone.DateAt(*h.ArchivedAt, loc).Prev(); lastActive.Before(end) {
			end = lastActive
		}
	}
	if end.Before(start) {
		return 0
	}
	return daysInRange(start, end)
}

func (s *service) TaskThroughput(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) (TaskThroughputReport, error) {
	if verr := validateRange(from, to); verr != nil {
		return TaskThroughputReport{}, verr
	}
	loc, err := s.zone.Zone(ctx, accountID)
	if err != nil {
		return TaskThroughputReport{}, fmt.Errorf("resolve account zone: %w", err)
	}
	start, _ := timezone.DayWindow(from, loc)
	_, end := timezone.DayWindow(to, loc)

	n, err := s.tasks.DoneCountInRange(ctx, accountID, start, end)
	if err != nil {
		return TaskThroughputReport{}, fmt.Errorf("done count in range: %w", err)
	}
	return TaskThroughputReport{From: from, To: to, DoneCount: n}, nil
}

func (s *service) DailyActualTotals(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) ([]timeline.DayTotal, error) {
	if verr := validateRange(from, to); verr != nil {
		return nil, verr
	}
	totals, err := s.timeline.DailyActualTotals(ctx, accountID, from, to)
	if err != nil {
		return nil, fmt.Errorf("daily actual totals: %w", err)
	}
	return totals, nil
}

func (s *service) Report(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) (Report, error) {
	if verr := validateReportRange(from, to); verr != nil {
		return Report{}, verr
	}

	timeByCategory, err := s.TimeByCategory(ctx, accountID, from, to)
	if err != nil {
		return Report{}, err
	}

	plannedVsActualAll, err := s.PlannedVsActualByCategory(ctx, accountID, from, to)
	if err != nil {
		return Report{}, err
	}
	plannedVsActual := make([]timeline.CategoryTotals, 0, len(plannedVsActualAll))
	for _, c := range plannedVsActualAll {
		if c.CategoryID != nil { // Uncategorized excluded — planned time is only meaningful per named category
			plannedVsActual = append(plannedVsActual, c)
		}
	}

	habitCompletion, err := s.HabitCompletion(ctx, accountID, from, to)
	if err != nil {
		return Report{}, err
	}

	throughput, err := s.TaskThroughput(ctx, accountID, from, to)
	if err != nil {
		return Report{}, err
	}

	dailyTotals, err := s.DailyActualTotals(ctx, accountID, from, to)
	if err != nil {
		return Report{}, err
	}

	return Report{
		From: from, To: to,
		TimeByCategory:    timeByCategory,
		PlannedVsActual:   plannedVsActual,
		HabitCompletion:   habitCompletion,
		TaskThroughput:    throughput.DoneCount,
		DailyActualTotals: dailyTotals,
	}, nil
}

func validateReportRange(from, to timezone.Date) *ValidationError {
	if verr := validateRange(from, to); verr != nil {
		return verr
	}
	if daysInRange(from, to) > maxReportRangeDays {
		return &ValidationError{Fields: map[string]string{
			"range": fmt.Sprintf("must not exceed %d days", maxReportRangeDays),
		}}
	}
	return nil
}

// daysInRange is the inclusive day count of [from, to] — always >= 1 since
// validateRange has already rejected to < from.
func daysInRange(from, to timezone.Date) int {
	n := 1
	for d := from; d.Before(to); d = d.Next() {
		n++
	}
	return n
}
