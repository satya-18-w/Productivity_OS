package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/categories"
	"github.com/satya-18-w/productivity-os/internal/goals"
	"github.com/satya-18-w/productivity-os/internal/habits"
	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
	"github.com/satya-18-w/productivity-os/internal/tasks"
	"github.com/satya-18-w/productivity-os/internal/timeline"
)

type fakeZone struct{}

func (fakeZone) Zone(context.Context, uuid.UUID) (*time.Location, error) { return time.UTC, nil }

// TestCategoriesOverview exercises the cmd/server composition handler end to end
// against a real database — every other module's tests cover their own
// CountByCategory; this proves the assembly (ADR-0009).
func TestCategoriesOverview(t *testing.T) {
	pool := pgtest.Pool(t)

	var acc uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO accounts (email, password_hash, timezone) VALUES ($1, 'x', 'UTC') RETURNING id`,
		"overview@test").Scan(&acc))

	catSvc := categories.NewService(pool)
	goalsSvc := goals.NewService(pool, catSvc)
	tasksSvc := tasks.NewService(pool, catSvc, goalsSvc)
	habitsSvc := habits.NewService(pool, fakeZone{}, catSvc)
	timelineSvc := timeline.NewService(pool, fakeZone{}, catSvc, tasksSvc)

	ctx := reqctx.WithIdentity(context.Background(), reqctx.Identity{AccountID: acc})

	deepWork, err := catSvc.Create(ctx, acc, categories.Input{Name: "Deep Work", Colour: "blue", Icon: "brain"})
	require.NoError(t, err)
	_, err = catSvc.Create(ctx, acc, categories.Input{Name: "Empty"})
	require.NoError(t, err)

	_, err = tasksSvc.CreateTask(ctx, acc, tasks.TaskInput{Title: "t1", CategoryID: &deepWork.ID})
	require.NoError(t, err)
	_, err = tasksSvc.CreateTask(ctx, acc, tasks.TaskInput{Title: "t2", CategoryID: &deepWork.ID})
	require.NoError(t, err)
	_, err = habitsSvc.CreateHabit(ctx, acc, habits.HabitInput{Name: "h1", CategoryID: &deepWork.ID})
	require.NoError(t, err)
	_, err = goalsSvc.CreateGoal(ctx, acc, goals.GoalInput{Title: "g1", CategoryID: &deepWork.ID})
	require.NoError(t, err)
	_, err = timelineSvc.AddBlock(ctx, acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: time.Date(2025, 6, 15, 9, 0, 0, 0, time.UTC),
		EndsAt: time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC), CategoryID: &deepWork.ID,
	})
	require.NoError(t, err)

	mux := http.NewServeMux()
	read := func(fn http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn(w, r.WithContext(reqctx.WithIdentity(r.Context(), reqctx.Identity{AccountID: acc})))
		})
	}
	mux.Handle("GET /api/categories/overview",
		read(categoriesOverviewHandler(catSvc, tasksSvc, habitsSvc, goalsSvc, timelineSvc)))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/categories/overview", nil)
	require.NoError(t, err)
	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body struct {
		Categories []categoryOverviewRow `json:"categories"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Len(t, body.Categories, 2)

	byName := map[string]categoryOverviewRow{}
	for _, c := range body.Categories {
		byName[c.Name] = c
	}

	dw := byName["Deep Work"]
	require.Equal(t, "blue", dw.Colour)
	require.Equal(t, "brain", dw.Icon)
	require.Equal(t, categoryCountsBody{Tasks: 2, Habits: 1, Goals: 1, Blocks: 1}, dw.Counts)

	e := byName["Empty"]
	require.Equal(t, categoryCountsBody{}, e.Counts, "a category with no items shows zero counts, not an error")
}
