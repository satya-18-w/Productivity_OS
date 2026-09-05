// Command server is the single Productivity OS application process. It serves the
// JSON API under /api and (from Phase 8) the built web client, from one origin.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/account"
	"github.com/satya-18-w/productivity-os/internal/categories"
	"github.com/satya-18-w/productivity-os/internal/export"
	"github.com/satya-18-w/productivity-os/internal/goals"
	"github.com/satya-18-w/productivity-os/internal/habits"
	"github.com/satya-18-w/productivity-os/internal/notes"
	"github.com/satya-18-w/productivity-os/internal/platform/config"
	"github.com/satya-18-w/productivity-os/internal/platform/httpx"
	"github.com/satya-18-w/productivity-os/internal/platform/postgres"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
	"github.com/satya-18-w/productivity-os/internal/reports"
	"github.com/satya-18-w/productivity-os/internal/reviews"
	"github.com/satya-18-w/productivity-os/internal/tasks"
	"github.com/satya-18-w/productivity-os/internal/timeline"
	"github.com/satya-18-w/productivity-os/web"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	setupLogger(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	slog.Info("database connected")

	accountSvc := account.NewService(pool, cfg.SessionTTL)
	accountHandler := account.NewHandler(accountSvc, cfg.IsProduction())

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if !postgres.Healthy(r.Context(), pool) {
			httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	accountHandler.MountPublic(mux)
	accountHandler.MountAuthed(mux)

	write := func(fn http.HandlerFunc) http.Handler {
		return accountHandler.RequireAuth(accountHandler.RequireCSRF(fn))
	}
	read := func(fn http.HandlerFunc) http.Handler {
		return accountHandler.RequireAuth(fn)
	}
	zone := accountZone{accountSvc}
	categorySvc := categories.NewService(pool)
	categories.NewHandler(categorySvc).Mount(mux, write, read)

	goalsSvc := goals.NewService(pool, categorySvc)
	tasksSvc := tasks.NewService(pool, categorySvc, goalsSvc)
	goals.NewHandler(goalsSvc, tasksSvc).Mount(mux, write, read)
	tasks.NewHandler(tasksSvc, zone).Mount(mux, write, read)

	timelineSvc := timeline.NewService(pool, zone, categorySvc, tasksSvc)
	timeline.NewHandler(timelineSvc, zone).Mount(mux, write, read)

	habitsSvc := habits.NewService(pool, zone, categorySvc)
	habits.NewHandler(habitsSvc, zone).Mount(mux, write, read)

	reviewsSvc := reviews.NewService(pool)
	reviews.NewHandler(reviewsSvc).Mount(mux, write, read)

	notesSvc := notes.NewService(pool)
	notes.NewHandler(notesSvc).Mount(mux, write, read)

	reportsSvc := reports.NewService(timelineSvc, habitsSvc, tasksSvc, zone)
	reports.NewHandler(reportsSvc).Mount(mux, read)

	exportSvc := export.NewService(categorySvc, timelineSvc, tasksSvc, habitsSvc, goalsSvc, reviewsSvc, notesSvc)
	export.NewHandler(exportSvc).Mount(mux, read)

	mux.Handle("GET /api/categories/overview",
		read(categoriesOverviewHandler(categorySvc, tasksSvc, habitsSvc, goalsSvc, timelineSvc)))

	// Everything not matched above is the frontend (client-side routing).
	if bundle, ok := web.Bundle(); ok {
		mux.Handle("/", httpx.SPA(bundle))
		slog.Info("serving embedded frontend bundle")
	} else {
		slog.Warn("no frontend bundle embedded; run `make web-build` before `go build`")
	}

	handler := httpx.Chain(mux,
		httpx.RequestIDMiddleware,
		httpx.Logger,
		httpx.Recoverer,
	)

	srv := httpx.NewServer(":"+cfg.Port, handler, cfg.ShutdownGrace)
	return srv.Run(ctx)
}

// accountZone adapts the account module to timeline.AccountZone.
type accountZone struct{ svc account.Service }

func (a accountZone) Zone(ctx context.Context, id uuid.UUID) (*time.Location, error) {
	p, err := a.svc.Read(ctx, id)
	if err != nil {
		return nil, err
	}
	return timezone.LoadLocation(p.Timezone)
}

func setupLogger(cfg config.Config) {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	var h slog.Handler = slog.NewJSONHandler(os.Stdout, opts)
	if !cfg.IsProduction() {
		opts.Level = slog.LevelDebug
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(h))
}
