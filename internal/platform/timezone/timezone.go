// Package timezone is the one place calendar reasoning happens (ADR-0005, N4).
// Every "which date/week is this instant in" and "what instant range does this
// date/week cover in the account's zone" question is answered here — never with
// AT TIME ZONE scattered through SQL or the server's local clock.
package timezone

import (
	"fmt"
	"sync"
	"time"

	_ "time/tzdata" // embed the zoneinfo DB so behaviour does not depend on the host
)

var locCache sync.Map // name -> *time.Location

// LoadLocation resolves an IANA name to a *time.Location, cached.
func LoadLocation(name string) (*time.Location, error) {
	if v, ok := locCache.Load(name); ok {
		return v.(*time.Location), nil
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, err
	}
	locCache.Store(name, loc)
	return loc, nil
}

// Valid reports whether name is a loadable IANA time-zone name. The bare "Local"
// and "" are rejected — an account must carry a real zone.
func Valid(name string) bool {
	if name == "" || name == "Local" {
		return false
	}
	_, err := LoadLocation(name)
	return err == nil
}

// Date is a calendar date with no time-of-day and no zone.
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

// ParseDate reads a YYYY-MM-DD string.
func ParseDate(s string) (Date, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return Date{}, fmt.Errorf("invalid date %q: want YYYY-MM-DD", s)
	}
	return Date{t.Year(), t.Month(), t.Day()}, nil
}

// String renders the date as YYYY-MM-DD.
func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, int(d.Month), d.Day)
}

// IsZero reports whether d is the zero Date.
func (d Date) IsZero() bool { return d == Date{} }

// Next returns the following calendar date.
func (d Date) Next() Date {
	t := time.Date(d.Year, d.Month, d.Day+1, 0, 0, 0, 0, time.UTC)
	return Date{t.Year(), t.Month(), t.Day()}
}

// Prev returns the preceding calendar date.
func (d Date) Prev() Date {
	t := time.Date(d.Year, d.Month, d.Day-1, 0, 0, 0, 0, time.UTC)
	return Date{t.Year(), t.Month(), t.Day()}
}

// AddDays returns the date n days from d (n may be negative).
func (d Date) AddDays(n int) Date {
	t := time.Date(d.Year, d.Month, d.Day+n, 0, 0, 0, 0, time.UTC)
	return Date{t.Year(), t.Month(), t.Day()}
}

// Before reports whether d is an earlier calendar date than o.
func (d Date) Before(o Date) bool {
	if d.Year != o.Year {
		return d.Year < o.Year
	}
	if d.Month != o.Month {
		return d.Month < o.Month
	}
	return d.Day < o.Day
}

// Today is the current calendar date in loc.
func Today(loc *time.Location) Date { return DateAt(time.Now(), loc) }

// DateAt is the calendar date that instant falls on when read in loc.
func DateAt(instant time.Time, loc *time.Location) Date {
	t := instant.In(loc)
	return Date{t.Year(), t.Month(), t.Day()}
}

// DayWindow is the half-open instant range [start, end) covering the calendar
// date d in loc. On a DST-transition day this range is 23h, 25h, or an odd
// fraction — that is correct, and callers must use the returned instants, not
// assume 24h.
func DayWindow(d Date, loc *time.Location) (start, end time.Time) {
	start = time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, loc)
	end = time.Date(d.Year, d.Month, d.Day+1, 0, 0, 0, 0, loc)
	return start, end
}

// ISOWeekWindow is the half-open instant range [Monday 00:00, next Monday 00:00)
// for the ISO week containing date d, in loc.
func ISOWeekWindow(d Date, loc *time.Location) (start, end time.Time) {
	midnight := time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, loc)
	daysSinceMonday := (int(midnight.Weekday()) + 6) % 7 // Go: Sun=0 → treat Mon as 0
	start = time.Date(d.Year, d.Month, d.Day-daysSinceMonday, 0, 0, 0, 0, loc)
	end = time.Date(start.Year(), start.Month(), start.Day()+7, 0, 0, 0, 0, loc)
	return start, end
}

// ISOWeekAt is the ISO year and week number for instant, read in loc.
func ISOWeekAt(instant time.Time, loc *time.Location) (year, week int) {
	return instant.In(loc).ISOWeek()
}

// OverlapSeconds is the number of seconds of the half-open block [bStart, bEnd)
// that fall inside the half-open window [wStart, wEnd). It is 0 when they do not
// overlap. This is the primitive for per-date totals of midnight-spanning blocks.
func OverlapSeconds(bStart, bEnd, wStart, wEnd time.Time) float64 {
	lo := bStart
	if wStart.After(lo) {
		lo = wStart
	}
	hi := bEnd
	if wEnd.Before(hi) {
		hi = wEnd
	}
	if !hi.After(lo) {
		return 0
	}
	return hi.Sub(lo).Seconds()
}
