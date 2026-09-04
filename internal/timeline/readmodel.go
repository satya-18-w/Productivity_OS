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

	out := DayTimeline{Date: date, Planned: []PositionedBlock{}, Actual: []PositionedBlock{}}
	for _, b := range blocks {
		pb := position(b, loc, date)
		if b.Kind == Planned {
			out.Planned = append(out.Planned, pb)
		} else {
			out.Actual = append(out.Actual, pb)
		}
	}
	return out, nil
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

	blocks, err := s.blocksOverlapping(ctx, accountID, start, end)
	if err != nil {
		return DayComparison{}, err
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
		// Clip the block to the date's window (N4: midnight-spanning correctness).
		secs := timezone.OverlapSeconds(b.StartsAt, b.EndsAt, start, end)
		if secs == 0 {
			continue
		}

		key, name := "", uncategorized
		if b.CategoryID != nil {
			key = b.CategoryID.String()
			if b.CategoryName != nil {
				name = *b.CategoryName
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

	result := DayComparison{Date: date, Categories: []CategoryTotals{}}
	for _, key := range order {
		g := byKey[key]
		p := int64(math.Round(g.planned))
		a := int64(math.Round(g.actual))
		result.Categories = append(result.Categories, CategoryTotals{
			CategoryID:        g.id,
			CategoryName:      g.name,
			PlannedSeconds:    p,
			ActualSeconds:     a,
			DifferenceSeconds: a - p,
		})
	}
	return result, nil
}

func (s *service) blocksOverlapping(ctx context.Context, accountID uuid.UUID, start, end time.Time) ([]Block, error) {
	rows, err := s.q.ListBlocksOverlapping(ctx, timelinedb.ListBlocksOverlappingParams{
		AccountID:   accountID,
		WindowStart: pgtype.Timestamptz{Time: start, Valid: true},
		WindowEnd:   pgtype.Timestamptz{Time: end, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("list blocks: %w", err)
	}

	out := make([]Block, len(rows))
	for i, r := range rows {
		out[i] = Block{
			ID:           r.ID,
			Kind:         BlockKind(r.Kind),
			StartsAt:     r.StartsAt.Time,
			EndsAt:       r.EndsAt.Time,
			CategoryID:   fromPgUUID(r.CategoryID),
			CategoryName: fromPgText(r.CategoryName),
		}
	}
	return out, nil
}
