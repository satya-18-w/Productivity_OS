// Command migrate applies the embedded SQL migrations. Used by `make migrate` for
// local development and as the one-shot step before server startup in a deployed
// environment (ADR-0003).
//
// Usage: migrate [up|down|drop]   (default: up)
package main

import (
	"fmt"
	"os"

	"github.com/satya-18-w/productivity-os/internal/platform/config"
	"github.com/satya-18-w/productivity-os/internal/platform/postgres"
)

func main() {
	cmd := "up"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	cfg, err := config.Load()
	if err != nil {
		fail(err)
	}

	switch cmd {
	case "up":
		err = postgres.MigrateUp(cfg.DatabaseURL)
	case "down":
		err = postgres.MigrateDown(cfg.DatabaseURL)
	case "drop":
		err = postgres.MigrateDrop(cfg.DatabaseURL)
	default:
		fail(fmt.Errorf("unknown command %q (want up|down|drop)", cmd))
	}
	if err != nil {
		fail(err)
	}
	fmt.Printf("migrate %s: ok\n", cmd)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "migrate:", err)
	os.Exit(1)
}
