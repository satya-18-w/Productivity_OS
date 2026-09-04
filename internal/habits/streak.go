package habits

import "github.com/satya-18-w/productivity-os/internal/platform/timezone"

// currentStreak counts consecutive completed dates ending at today or yesterday
// (v1.md §9). completed is the set of completed dates (YYYY-MM-DD keys).
// Completions after today do not count toward the current streak.
func currentStreak(completed map[string]struct{}, today timezone.Date) int {
	_, hasToday := completed[today.String()]
	anchor := today
	if !hasToday {
		anchor = today.Prev()
		if _, hasYesterday := completed[anchor.String()]; !hasYesterday {
			return 0
		}
	}

	n := 0
	for d := anchor; ; d = d.Prev() {
		if _, ok := completed[d.String()]; !ok {
			break
		}
		n++
	}
	return n
}

// countInRange returns how many completed dates fall in [from, to] inclusive.
func countInRange(completed map[string]struct{}, from, to timezone.Date) int {
	n := 0
	for d := from; !to.Before(d); d = d.Next() {
		if _, ok := completed[d.String()]; ok {
			n++
		}
	}
	return n
}
