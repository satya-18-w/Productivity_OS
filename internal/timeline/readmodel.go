package timeline

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
	"github.com/satya-18-w/productivity-os/internal/timeline/timelinedb"
)

const uncategorized = "Uncategorized"

// maxRangeDays bounds TimelineRange — enough for a month grid's 42 visible cells
// (leading/trailing days from adjacent months included).
const maxRangeDays = 62

func (s *service) Timeline(ctx context.Context, accountID uuid.UUID, date timezone.Date) (DayTimeline, error) {
	loc, err := s.zone.Zone(ctx, accountID)
	if err != nil {
		return DayTimeline{}, fmt.Errorf("resolve account zone: %w", err)
	}
	start, end := timezone.DayWindow(date, loc)

	blocks, err := s.blocksOverlapping(ctx, accountID, start, end)
	if err != nil {
		return DayTimeline{}, err
	}

	names, err := s.cats.NamesForAccount(ctx, accountID)
	if err != nil {
		return DayTimeline{}, fmt.Errorf("resolve category names: %w", err)
	}

	out := DayTimeline{Date: date, Planned: []PositionedBlock{}, Actual: []PositionedBlock{}}
	for _, b := range blocks {
		if b.CategoryID != nil {
			if n, ok := names[*b.CategoryID]; ok {
				b.CategoryName = &n
			}
		}
		pb := position(b, loc, date)
		if b.Kind == Planned {
			out.Planned = append(out.Planned, pb)
		} else {
			out.Actual = append(out.Actual, pb)
		}
	}
	return out, nil
}

// TimelineRange batches Timeline over every day in [from, to] inclusive — one
// blocksOverlapping query and one category-name lookup for the whole window,
// rather than one of each per day, for the Week/Month views (MX5-range).
func (s *service) TimelineRange(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) (RangeTimeline, error) {
	if to.Before(from) {
		return RangeTimeline{}, &ValidationError{Fields: map[string]string{"to": "must not be before from"}}
	}
	if daysInRange(from, to) > maxRangeDays {
		return RangeTimeline{}, &ValidationError{Fields: map[string]string{"to": fmt.Sprintf("range must be at most %d days", maxRangeDays)}}
	}

	loc, err := s.zone.Zone(ctx, accountID)
	if err != nil {
		return RangeTimeline{}, fmt.Errorf("resolve account zone: %w", err)
	}
	start, _ := timezone.DayWindow(from, loc)
	_, end := timezone.DayWindow(to, loc)

	blocks, err := s.blocksOverlapping(ctx, accountID, start, end)
	if err != nil {
		return RangeTimeline{}, err
	}

	names, err := s.cats.NamesForAccount(ctx, accountID)
	if err != nil {
		return RangeTimeline{}, fmt.Errorf("resolve category names: %w", err)
	}
	for i := range blocks {
		if blocks[i].CategoryID != nil {
			if n, ok := names[*blocks[i].CategoryID]; ok {
				blocks[i].CategoryName = &n
			}
		}
	}

	days := make([]DayTimeline, 0, daysInRange(from, to))
	for d := from; !to.Before(d); d = d.Next() {
		dayStart, dayEnd := timezone.DayWindow(d, loc)
		dt := DayTimeline{Date: d, Planned: []PositionedBlock{}, Actual: []PositionedBlock{}}
		for _, b := range blocks {
			if !b.StartsAt.Before(dayEnd) || !b.EndsAt.After(dayStart) {
				continue // doesn't overlap this particular day
			}
			pb := position(b, loc, d)
			if b.Kind == Planned {
				dt.Planned = append(dt.Planned, pb)
			} else {
				dt.Actual = append(dt.Actual, pb)
			}
		}
		days = append(days, dt)
	}
	return RangeTimeline{From: from, To: to, Days: days}, nil
}

// daysInRange is the inclusive day count of [from, to] — always >= 1 since callers
// have already rejected to < from.
func daysInRange(from, to timezone.Date) int {
	n := 1
	for d := from; d.Before(to); d = d.Next() {
		n++
	}
	return n
}

// position places a block on the queried date's 24-hour grid using its
// wall-clock time in loc.
func position(b Block, loc *time.Location, day timezone.Date) PositionedBlock {
	startMin, fromPrev := minuteOfDay(b.StartsAt, loc, day, false)
	endMin, toNext := minuteOfDay(b.EndsAt, loc, day, true)

	ls := b.StartsAt.In(loc)
	le := b.EndsAt.In(loc)
	startDate := timezone.Date{Year: ls.Year(), Month: ls.Month(), Day: ls.Day()}
	endDate := timezone.Date{Year: le.Year(), Month: le.Month(), Day: le.Day()}

	return PositionedBlock{
		Block:       b,
		StartMinute: startMin,
		EndMinute:   endMin,
		FromPrevDay: fromPrev,
		ToNextDay:   toNext,
		LocalDate:   startDate.String(),
		LocalStart:  ls.Format("15:04"),
		LocalEnd:    le.Format("15:04"),
		EndsNextDay: startDate.Before(endDate),
	}
}

func minuteOfDay(instant time.Time, loc *time.Location, day timezone.Date, isEnd bool) (minute int, spill bool) {
	local := instant.In(loc)
	d := timezone.Date{Year: local.Year(), Month: local.Month(), Day: local.Day()}

	switch {
	case d.Before(day):
		return 0, true
	case day.Before(d):
		// An end exactly at the next midnight is the clean bottom of the grid,
		// not a spill.
		if isEnd && d == day.Next() && local.Hour() == 0 && local.Minute() == 0 {
			return 1440, false
		}
		return 1440, true
	default:
		return local.Hour()*60 + local.Minute(), false
	}
}

func (s *service) Comparison(ctx context.Context, accountID uuid.UUID, date timezone.Date) (DayComparison, error) {
	loc, err := s.zone.Zone(ctx, accountID)
	if err != nil {
		return DayComparison{}, fmt.Errorf("resolve account zone: %w", err)
	}
	start, end := timezone.DayWindow(date, loc)

	cats, err := s.categoryTotals(ctx, accountID, start, end)
	if err != nil {
		return DayComparison{}, err
	}
	return DayComparison{Date: date, Categories: cats}, nil
}

// ComparisonRange is the M6/M7 foundation: per-category planned/actual/difference
// totals summed over every day in [from, to] inclusive, in the account's timezone.
// Reuses Comparison's own bucketing logic over a wider window (M6 Phase 1).
func (s *service) ComparisonRange(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) (RangeComparison, error) {
	if to.Before(from) {
		return RangeComparison{}, &ValidationError{Fields: map[string]string{"to": "must not be before from"}}
	}
	loc, err := s.zone.Zone(ctx, accountID)
	if err != nil {
		return RangeComparison{}, fmt.Errorf("resolve account zone: %w", err)
	}
	start, _ := timezone.DayWindow(from, loc)
	_, end := timezone.DayWindow(to, loc)

	cats, err := s.categoryTotals(ctx, accountID, start, end)
	if err != nil {
		return RangeComparison{}, err
	}
	return RangeComparison{From: from, To: to, Categories: cats}, nil
}

// DailyActualTotals is the M7 "Daily actual totals" report: total actual (not
// planned) time for each day in [from, to] inclusive, in the account's timezone.
func (s *service) DailyActualTotals(ctx context.Context, accountID uuid.UUID, from, to timezone.Date) ([]DayTotal, error) {
	if to.Before(from) {
		return nil, &ValidationError{Fields: map[string]string{"to": "must not be before from"}}
	}
	loc, err := s.zone.Zone(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("resolve account zone: %w", err)
	}
	start, _ := timezone.DayWindow(from, loc)
	_, end := timezone.DayWindow(to, loc)

	blocks, err := s.blocksOverlapping(ctx, accountID, start, end)
	if err != nil {
		return nil, err
	}

	out := []DayTotal{}
	for d := from; !to.Before(d); d = d.Next() {
		dayStart, dayEnd := timezone.DayWindow(d, loc)
		var total float64
		for _, b := range blocks {
			if b.Kind != Actual {
				continue
			}
			total += timezone.OverlapSeconds(b.StartsAt, b.EndsAt, dayStart, dayEnd)
		}
		out = append(out, DayTotal{Date: d, ActualSeconds: int64(math.Round(total))})
	}
	return out, nil
}

// categoryTotals sums planned/actual seconds per category for every block
// overlapping [start, end), clipped to that window (N4: midnight-spanning
// correctness holds for a single day and for a multi-day range alike).
func (s *service) categoryTotals(ctx context.Context, accountID uuid.UUID, start, end time.Time) ([]CategoryTotals, error) {
	blocks, err := s.blocksOverlapping(ctx, accountID, start, end)
	if err != nil {
		return nil, err
	}

	names, err := s.cats.NamesForAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("resolve category names: %w", err)
	}

	type bucket struct {
		id      *uuid.UUID
		name    string
		planned float64
		actual  float64
	}
	byKey := map[string]*bucket{}
	var order []string

	for _, b := range blocks {
		secs := timezone.OverlapSeconds(b.StartsAt, b.EndsAt, start, end)
		if secs == 0 {
			continue
		}

		key, name := "", uncategorized
		if b.CategoryID != nil {
			key = b.CategoryID.String()
			if n, ok := names[*b.CategoryID]; ok {
				name = n
			}
		}

		g, ok := byKey[key]
		if !ok {
			g = &bucket{id: b.CategoryID, name: name}
			byKey[key] = g
			order = append(order, key)
		}
		if b.Kind == Planned {
			g.planned += secs
		} else {
			g.actual += secs
		}
	}

	// Named categories by name; the Uncategorized bucket ("" key) last.
	sort.SliceStable(order, func(i, j int) bool {
		if order[i] == "" {
			return false
		}
		if order[j] == "" {
			return true
		}
		return byKey[order[i]].name < byKey[order[j]].name
	})

	result := []CategoryTotals{}
	for _, key := range order {
		g := byKey[key]
		p := int64(math.Round(g.planned))
		a := int64(math.Round(g.actual))
		result = append(result, CategoryTotals{
			CategoryID:        g.id,
			CategoryName:      g.name,
			PlannedSeconds:    p,
			ActualSeconds:     a,
			DifferenceSeconds: a - p,
		})
	}
	return result, nil
}

// CountByCategory implements categories.Counter for the categories overview
// (ADR-0009). A task-linked block never carries its own category_id (the MX-TL
// CHECK constraint), so its contribution is counted separately via CountBlocksByTask
// and attributed to its linked task's category — otherwise it would silently
// disappear from every category's block total.
func (s *service) CountByCategory(ctx context.Context, accountID uuid.UUID) (map[uuid.UUID]int, error) {
	rows, err := s.q.CountBlocksByCategory(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("count blocks by category: %w", err)
	}
	out := make(map[uuid.UUID]int, len(rows))
	for _, r := range rows {
		if id := fromPgUUID(r.CategoryID); id != nil {
			out[*id] = int(r.Total)
		}
	}

	taskRows, err := s.q.CountBlocksByTask(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("count blocks by task: %w", err)
	}
	if len(taskRows) == 0 {
		return out, nil
	}
	taskCats, err := s.tasks.CategoriesForTasks(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("resolve task categories: %w", err)
	}
	for _, r := range taskRows {
		taskID := fromPgUUID(r.TaskID)
		if taskID == nil {
			continue
		}
		if catID, ok := taskCats[*taskID]; ok && catID != nil {
			out[*catID] += int(r.Total)
		}
	}
	return out, nil
}

// blocksOverlapping is the shared read path behind Timeline, Comparison,
// ComparisonRange, and DailyActualTotals. It resolves each task-linked block's
// inherited category here, once, so every consumer sees the effective category
// without knowing the inheritance rule exists (MX-TL).
func (s *service) blocksOverlapping(ctx context.Context, accountID uuid.UUID, start, end time.Time) ([]Block, error) {
	rows, err := s.q.ListBlocksOverlapping(ctx, timelinedb.ListBlocksOverlappingParams{
		AccountID:   accountID,
		WindowStart: pgtype.Timestamptz{Time: start, Valid: true},
		WindowEnd:   pgtype.Timestamptz{Time: end, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("list blocks: %w", err)
	}

	taskCats, err := s.taskCategoriesIfNeeded(ctx, accountID, func() bool {
		for _, r := range rows {
			if r.TaskID.Valid {
				return true
			}
		}
		return false
	}())
	if err != nil {
		return nil, err
	}

	out := make([]Block, len(rows))
	for i, r := range rows {
		out[i] = blockFromRow(r.ID, r.Kind, r.StartsAt, r.EndsAt, r.CategoryID, r.TaskID, taskCats)
	}
	return out, nil
}

// BlocksForTask returns every block linked to taskID, across any date, with the
// task's own (inherited) category resolved once for the whole list — every block
// returned shares the same category, since it is the task's, not the block's own
// (v1.md §7, MX-TL's inheritance rule).
func (s *service) BlocksForTask(ctx context.Context, accountID, taskID uuid.UUID) ([]Block, error) {
	ok, err := s.tasks.AssignableToAccount(ctx, accountID, taskID)
	if err != nil {
		return nil, fmt.Errorf("check task: %w", err)
	}
	if !ok {
		return nil, ErrTaskNotFound
	}

	rows, err := s.q.ListBlocksByTask(ctx, timelinedb.ListBlocksByTaskParams{
		AccountID: accountID,
		TaskID:    pgtype.UUID{Bytes: taskID, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("list blocks for task: %w", err)
	}
	if len(rows) == 0 {
		return []Block{}, nil
	}

	taskCats, err := s.tasks.CategoriesForTasks(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("resolve task categories: %w", err)
	}
	categoryID := taskCats[taskID]

	var categoryName *string
	if categoryID != nil {
		names, err := s.cats.NamesForAccount(ctx, accountID)
		if err != nil {
			return nil, fmt.Errorf("resolve category names: %w", err)
		}
		if n, ok := names[*categoryID]; ok {
			categoryName = &n
		}
	}

	out := make([]Block, len(rows))
	for i, r := range rows {
		out[i] = Block{
			ID:           r.ID,
			Kind:         BlockKind(r.Kind),
			StartsAt:     r.StartsAt.Time,
			EndsAt:       r.EndsAt.Time,
			TaskID:       fromPgUUID(r.TaskID),
			CategoryID:   categoryID,
			CategoryName: categoryName,
		}
	}
	return out, nil
}

func (s *service) ListAllBlocks(ctx context.Context, accountID uuid.UUID) ([]Block, error) {
	rows, err := s.q.ListAllBlocks(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("list all blocks: %w", err)
	}
	out := make([]Block, len(rows))
	for i, r := range rows {
		out[i] = Block{
			ID:         r.ID,
			Kind:       BlockKind(r.Kind),
			StartsAt:   r.StartsAt.Time,
			EndsAt:     r.EndsAt.Time,
			CategoryID: fromPgUUID(r.CategoryID),
			TaskID:     fromPgUUID(r.TaskID),
		}
	}
	return out, nil
}

// taskCategoriesIfNeeded resolves the caller's task->category map only when at
// least one row actually needs it, avoiding an extra query for accounts with no
// task-linked blocks in the window.
func (s *service) taskCategoriesIfNeeded(ctx context.Context, accountID uuid.UUID, needed bool) (map[uuid.UUID]*uuid.UUID, error) {
	if !needed {
		return nil, nil
	}
	taskCats, err := s.tasks.CategoriesForTasks(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("resolve task categories: %w", err)
	}
	return taskCats, nil
}

// blockFromRow builds a Block from a raw row, resolving an inherited category when
// taskID is set (MX-TL) — categoryID is only meaningful when taskID is nil.
func blockFromRow(id uuid.UUID, kind string, startsAt, endsAt pgtype.Timestamptz, categoryID, taskID pgtype.UUID, taskCats map[uuid.UUID]*uuid.UUID) Block {
	b := Block{
		ID:       id,
		Kind:     BlockKind(kind),
		StartsAt: startsAt.Time,
		EndsAt:   endsAt.Time,
		TaskID:   fromPgUUID(taskID),
	}
	if b.TaskID != nil {
		b.CategoryID = taskCats[*b.TaskID]
	} else {
		b.CategoryID = fromPgUUID(categoryID)
	}
	return b
}
