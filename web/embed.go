// Package web embeds the built frontend bundle so the Go binary serves it at a
// single origin (ADR-0001, ADR-0006). The bundle is produced by `pnpm build` into
// web/dist; until then only the .gitkeep placeholder is embedded and Bundle
// reports that no frontend is available.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var embedded embed.FS

// Bundle returns the built frontend files rooted at dist/, and true if a real
// bundle (with an index.html) is present.
func Bundle() (fs.FS, bool) {
	sub, err := fs.Sub(embedded, "dist")
	if err != nil {
		return nil, false
	}
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		return nil, false
	}
	return sub, true
}
