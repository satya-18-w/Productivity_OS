package timezone_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

func loc(t *testing.T, name string) *time.Location {
	t.Helper()
	l, err := timezone.LoadLocation(name)
	require.NoError(t, err)
	return l
}

func TestValid(t *testing.T) {
	cases := map[string]bool{
		"Asia/Kolkata":      true,
		"America/New_York":  true,
		"Europe/London":     true,
		"UTC":               true,
		"":                  false,
		"Local":             false,
		"Mars/Olympus_Mons": false,
		"asia/kolkata":      false,
		"GMT+5":             false,
	}
	for name, want := range cases {
		require.Equalf(t, want, timezone.Valid(name), "Valid(%q)", name)
	}
}

func TestParseDate(t *testing.T) {
	d, err := timezone.ParseDate("2025-03-09")
	require.NoError(t, err)
	require.Equal(t, timezone.Date{Year: 2025, Month: time.March, Day: 9}, d)
	require.Equal(t, "2025-03-09", d.String())

	_, err = timezone.ParseDate("09/03/2025")
	require.Error(t, err)
	_, err = timezone.ParseDate("2025-13-01")
	require.Error(t, err)
}

func TestDayWindow_NormalDay(t *testing.T) {
	d := timezone.Date{Year: 2025, Month: time.June, Day: 15}
	for _, name := range []string{"Asia/Kolkata", "UTC", "Pacific/Chatham", "America/New_York"} {
		start, end := timezone.DayWindow(d, loc(t, name))
		require.Equalf(t, 24*time.Hour, end.Sub(start), "%s should be a 24h day", name)
		require.Equal(t, 0, start.In(loc(t, name)).Hour())
	}
}

func TestDayWindow_DSTTransitions(t *testing.T) {
	ny := loc(t, "America/New_York")

	// Spring forward: 2025-03-09, 02:00 -> 03:00, a 23-hour day.
	start, end := timezone.DayWindow(timezone.Date{Year: 2025, Month: time.March, Day: 9}, ny)
	require.Equal(t, 23*time.Hour, end.Sub(start))

	// Fall back: 2025-11-02, 02:00 -> 01:00, a 25-hour day.
	start, end = timezone.DayWindow(timezone.Date{Year: 2025, Month: time.November, Day: 2}, ny)
	require.Equal(t, 25*time.Hour, end.Sub(start))

	// Lord Howe Island shifts DST by only 30 minutes.
	lh := loc(t, "Australia/Lord_Howe")
	start, end = timezone.DayWindow(timezone.Date{Year: 2025, Month: time.October, Day: 5}, lh)
	require.Equal(t, 23*time.Hour+30*time.Minute, end.Sub(start))
	start, end = timezone.DayWindow(timezone.Date{Year: 2025, Month: time.April, Day: 6}, lh)
	require.Equal(t, 24*time.Hour+30*time.Minute, end.Sub(start))
}

func TestDayWindow_ChathamOffsetHasMinutes(t *testing.T) {
	// Chatham is +12:45 / +13:45 — a 45-minute offset. Local midnight must still
	// be a clean local 00:00, and the UTC instant lands on a :15 minute.
	ch := loc(t, "Pacific/Chatham")
	start, _ := timezone.DayWindow(timezone.Date{Year: 2025, Month: time.January, Day: 10}, ch)

	local := start.In(ch)
	require.Equal(t, 0, local.Hour())
	require.Equal(t, 0, local.Minute())

	_, offset := local.Zone()
	require.Equal(t, 45, (offset/60)%60, "Chatham offset carries 45 minutes")
}

func TestDateAt(t *testing.T) {
	ny := loc(t, "America/New_York")
	kol := loc(t, "Asia/Kolkata")

	// 2025-06-15 23:30 in New York is already 2025-06-16 in Kolkata.
	instant := time.Date(2025, time.June, 15, 23, 30, 0, 0, ny)
	require.Equal(t, timezone.Date{Year: 2025, Month: time.June, Day: 15}, timezone.DateAt(instant, ny))
	require.Equal(t, timezone.Date{Year: 2025, Month: time.June, Day: 16}, timezone.DateAt(instant, kol))
}

func TestOverlapSeconds_MidnightSpanningBlock(t *testing.T) {
	ny := loc(t, "America/New_York")
	dayN := timezone.Date{Year: 2025, Month: time.June, Day: 15}
	dayN1 := timezone.Date{Year: 2025, Month: time.June, Day: 16}

	// A block 23:00 (15th) -> 02:00 (16th), local time.
	bStart := time.Date(2025, time.June, 15, 23, 0, 0, 0, ny)
	bEnd := time.Date(2025, time.June, 16, 2, 0, 0, 0, ny)

	nStart, nEnd := timezone.DayWindow(dayN, ny)
	require.InDelta(t, 3600, timezone.OverlapSeconds(bStart, bEnd, nStart, nEnd), 0.5, "1h on day N")

	n1Start, n1End := timezone.DayWindow(dayN1, ny)
	require.InDelta(t, 7200, timezone.OverlapSeconds(bStart, bEnd, n1Start, n1End), 0.5, "2h on day N+1")
}

func TestOverlapSeconds(t *testing.T) {
	base := time.Date(2025, time.June, 15, 0, 0, 0, 0, time.UTC)
	h := func(n int) time.Time { return base.Add(time.Duration(n) * time.Hour) }

	require.InDelta(t, 3600, timezone.OverlapSeconds(h(23), h(26), h(0), h(24)), 0.5)  // 23:00-24:00
	require.InDelta(t, 7200, timezone.OverlapSeconds(h(23), h(26), h(24), h(48)), 0.5) // 00:00-02:00 next day
	require.Equal(t, 0.0, timezone.OverlapSeconds(h(1), h(2), h(5), h(6)))             // disjoint
	require.Equal(t, 0.0, timezone.OverlapSeconds(h(2), h(2), h(0), h(24)))            // empty block
	require.InDelta(t, 3600, timezone.OverlapSeconds(h(9), h(10), h(0), h(24)), 0.5)   // fully inside
}

func TestISOWeekWindow_AcrossYearBoundary(t *testing.T) {
	utc := time.UTC
	// 2021-01-01 is a Friday in ISO week 2020-W53; that week's Monday is 2020-12-28.
	start, end := timezone.ISOWeekWindow(timezone.Date{Year: 2021, Month: time.January, Day: 1}, utc)
	require.Equal(t, time.Date(2020, time.December, 28, 0, 0, 0, 0, utc), start)
	require.Equal(t, time.Date(2021, time.January, 4, 0, 0, 0, 0, utc), end)
	require.Equal(t, 7*24*time.Hour, end.Sub(start))

	y, w := timezone.ISOWeekAt(start, utc)
	require.Equal(t, 2020, y)
	require.Equal(t, 53, w)
}

func TestISOWeekWindow_MondayIsIdempotent(t *testing.T) {
	utc := time.UTC
	// 2025-06-16 is a Monday.
	start, _ := timezone.ISOWeekWindow(timezone.Date{Year: 2025, Month: time.June, Day: 16}, utc)
	require.Equal(t, time.Date(2025, time.June, 16, 0, 0, 0, 0, utc), start)
}
