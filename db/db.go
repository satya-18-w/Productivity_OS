// Package db holds the SQL migrations, embedded so the binary can apply them with
// no external files. Migrations are forward-only (ADR-0003): never edit one that
// has shipped — add a new one.
package db

import "embed"

// MigrationsFS contains every file under migrations/.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS
