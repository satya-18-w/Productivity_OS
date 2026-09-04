package timezone_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

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
