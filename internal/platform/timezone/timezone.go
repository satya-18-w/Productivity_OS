// Package timezone validates IANA time-zone names. Date and week bucketing (the
// rest of requirement N4) arrives in M2; M1 only needs name validation (ADR-0005).
package timezone

import "time"

// Valid reports whether name is a loadable IANA time-zone name (e.g.
// "Asia/Kolkata"). The bare "Local" and "" are rejected — an account must carry a
// real zone.
func Valid(name string) bool {
	if name == "" || name == "Local" {
		return false
	}
	_, err := time.LoadLocation(name)
	return err == nil
}
