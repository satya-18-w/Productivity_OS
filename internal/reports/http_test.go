package reports_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
	"github.com/satya-18-w/productivity-os/internal/reports"
	"github.com/satya-18-w/productivity-os/internal/tasks"
	"github.com/satya-18-w/productivity-os/internal/timeline"
)

// stubProtector injects a fixed identity, standing in for the account middleware.
func stubProtector(accountID uuid.UUID) reports.Protector {
	return func(fn http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn(w, r.WithContext(reqctx.WithIdentity(r.Context(), reqctx.Identity{AccountID: accountID})))
		})
	}
}

// httpEnv wires the reports handler over the real domain modules, mirroring
// cmd/server's own wiring.
type httpEnv struct {
	srv      *httptest.Server
	acc      uuid.UUID
	timeline timeline.Service
	tasks    tasks.Service
}

func httpSetup(t *testing.T) httpEnv {
	t.Helper()
	e := setup(t)
	mux := http.NewServeMux()
	reports.NewHandler(e.reports).Mount(mux, stubProtector(e.acc))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return httpEnv{srv: srv, acc: e.acc, timeline: e.timeline, tasks: e.tasks}
}

func do(t *testing.T, srv *httptest.Server, method, path string) (*http.Response, map[string]any) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, srv.URL+path, strings.NewReader(""))
	require.NoError(t, err)
	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var m map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&m)
	return resp, m
}

// TestReportEndpoint verifies the combined GET /api/reports response matches
// docs/left.md Phase 9's shape field-for-field.
func TestReportEndpoint(t *testing.T) {
	e := httpSetup(t)
	ctx := context.Background()

	_, err := e.timeline.AddBlock(ctx, e.acc, timeline.BlockInput{
		Kind: timeline.Planned, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T11:00:00Z"),
	})
	require.NoError(t, err)
	_, err = e.timeline.AddBlock(ctx, e.acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T12:00:00Z"),
	})
	require.NoError(t, err)
	task, err := e.tasks.CreateTask(ctx, e.acc, tasks.TaskInput{Title: "t"})
	require.NoError(t, err)
	require.NoError(t, e.tasks.MoveTask(ctx, e.acc, task.ID, tasks.Done))

	resp, body := do(t, e.srv, http.MethodGet, "/api/reports?from=2025-06-15&to=2025-06-15")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "2025-06-15", body["from"])
	require.Equal(t, "2025-06-15", body["to"])

	tbc := body["time_by_category"].([]any)
	require.Len(t, tbc, 1)
	tbc0 := tbc[0].(map[string]any)
	require.Nil(t, tbc0["category_id"])
	require.Equal(t, "Uncategorized", tbc0["category_name"])
	require.Equal(t, float64(10800), tbc0["seconds"], "actual only")

	pva := body["planned_vs_actual"].([]any)
	require.Empty(t, pva, "Uncategorized is excluded from planned_vs_actual")

	require.Empty(t, body["habit_completion"])

	// task moved to DONE today, not on 2025-06-15, so throughput for that date is 0
	require.Equal(t, float64(0), body["task_throughput"])

	days := body["daily_actual_totals"].([]any)
	require.Len(t, days, 1)
	require.Equal(t, "2025-06-15", days[0].(map[string]any)["date"])
	require.Equal(t, float64(10800), days[0].(map[string]any)["seconds"])
}

func TestReportEndpoint_RangeValidation(t *testing.T) {
	e := httpSetup(t)

	// missing range -> 400
	resp, _ := do(t, e.srv, http.MethodGet, "/api/reports")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// malformed date -> 400
	resp, _ = do(t, e.srv, http.MethodGet, "/api/reports?from=nope&to=2025-06-15")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// to before from -> 400
	resp, _ = do(t, e.srv, http.MethodGet, "/api/reports?from=2025-06-16&to=2025-06-15")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// range too large -> 400
	resp, _ = do(t, e.srv, http.MethodGet, "/api/reports?from=2024-01-01&to=2026-01-01")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestReportEndpoint_Isolation(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	other := newAccount(t, e.pool, "other-http@test")

	cat := mkCategory(t, e.pool, e.acc, "Mine")
	_, err := e.timeline.AddBlock(ctx, e.acc, timeline.BlockInput{
		Kind: timeline.Actual, StartsAt: ts("2025-06-15T09:00:00Z"), EndsAt: ts("2025-06-15T10:00:00Z"), CategoryID: &cat,
	})
	require.NoError(t, err)

	mux := http.NewServeMux()
	reports.NewHandler(e.reports).Mount(mux, stubProtector(other))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	resp, body := do(t, srv, http.MethodGet, "/api/reports?from=2025-06-15&to=2025-06-15")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Empty(t, body["time_by_category"], "other sees none of the owner's time")
}
