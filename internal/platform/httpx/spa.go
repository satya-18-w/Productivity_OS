package httpx

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// SPA serves a single-page application from fsys: real files are served directly,
// and any other path returns index.html so the client router can handle it.
func SPA(fsys fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(fsys))

	index, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		// No bundle: report clearly rather than 404 mysteriously.
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			WriteError(w, r, NewError(http.StatusNotFound, CodeNotFound, "No frontend bundle is built"))
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if clean != "" && fileExists(fsys, clean) {
			fileServer.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(index)
	})
}

func fileExists(fsys fs.FS, name string) bool {
	f, err := fsys.Open(name)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	return err == nil && !info.IsDir()
}
