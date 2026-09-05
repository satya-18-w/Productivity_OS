package timeline_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/categories"
	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
	"github.com/satya-18-w/productivity-os/internal/timeline"
)

// stubProtector injects a fixed identity, standing in for the account middleware.
func stubProtector(accountID uuid.UUID) timeline.Protector {
	return func(fn http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn(w, r.WithContext(reqctx.WithIdentity(r.Context(), reqctx.Identity{AccountID: accountID})))
		})
	}
}

func httpSetup(t *testing.T) (*httptest.Server, uuid.UUID) {
	t.Helper()
	srv, _, acc := httpSetupWithPool(t)
	return srv, acc
}

func httpSetupWithPool(t *testing.T) (*httptest.Server, *pgxpool.Pool, uuid.UUID) {
	t.Helper()
	pool := pgtest.Pool(t)
	acc := newAccount(t, pool, "http@test")
	mux := http.NewServeMux()
	z := fakeZone{}
	svc := timeline.NewService(pool, z, categories.NewService(pool), newTasksSvc(pool))
	timeline.NewHandler(svc, z).Mount(mux, stubProtector(acc), stubProtector(acc))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, pool, acc
}

func do(t *testing.T, srv *httptest.Server, method, path, body string) (*http.Response, map[string]any) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, srv.URL+path, strings.NewReader(body))
	require.NoError(t, err)
	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var m map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&m)
	return resp, m
}

func TestBlockEndpoints_WallClock(t *testing.T) {
	srv, _ := httpSetup(t) // fakeZone => UTC

	// create a planned block 09:00–11:00 on 2025-06-15
	resp, body := do(t, srv, http.MethodPost, "/api/blocks",
		`{"kind":"planned","date":"2025-06-15","start":"09:00","end":"11:00"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, "2025-06-15T09:00:00Z", body["starts_at"])
	require.Equal(t, "2025-06-15T11:00:00Z", body["ends_at"])

	// bad time -> 400
	resp, _ = do(t, srv, http.MethodPost, "/api/blocks",
		`{"kind":"actual","date":"2025-06-15","start":"9am","end":"11:00"}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// it shows on the timeline, positioned
	_, tl := do(t, srv, http.MethodGet, "/api/timeline?date=2025-06-15", "")
	planned := tl["planned"].([]any)
	require.Len(t, planned, 1)
	p0 := planned[0].(map[string]any)
	require.Equal(t, float64(540), p0["start_minute"])
	require.Equal(t, float64(660), p0["end_minute"])
	require.Equal(t, false, p0["from_prev_day"])
	require.Equal(t, false, p0["to_next_day"])
}

func TestBlockEndpoints_CrossMidnight(t *testing.T) {
	srv, _ := httpSetup(t)

	resp, _ := do(t, srv, http.MethodPost, "/api/blocks",
		`{"kind":"actual","date":"2025-06-15","start":"23:00","end":"01:00","ends_next_day":true}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	_, d15 := do(t, srv, http.MethodGet, "/api/timeline?date=2025-06-15", "")
	a15 := d15["actual"].([]any)[0].(map[string]any)
	require.Equal(t, float64(1380), a15["start_minute"])
	require.Equal(t, float64(1440), a15["end_minute"])
	require.Equal(t, true, a15["to_next_day"])

	_, d16 := do(t, srv, http.MethodGet, "/api/timeline?date=2025-06-16", "")
	a16 := d16["actual"].([]any)[0].(map[string]any)
	require.Equal(t, float64(0), a16["start_minute"])
	require.Equal(t, float64(60), a16["end_minute"])
	require.Equal(t, true, a16["from_prev_day"])
}

func TestTimelineEndpoint_MissingDate(t *testing.T) {
	srv, _ := httpSetup(t)
	resp, _ := do(t, srv, http.MethodGet, "/api/timeline", "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestTimelineRangeEndpoint(t *testing.T) {
	srv, _ := httpSetup(t)

	do(t, srv, http.MethodPost, "/api/blocks",
		`{"kind":"planned","date":"2025-06-15","start":"09:00","end":"10:00"}`)
	do(t, srv, http.MethodPost, "/api/blocks",
		`{"kind":"actual","date":"2025-06-17","start":"09:00","end":"10:00"}`)

	resp, body := do(t, srv, http.MethodGet, "/api/timeline/range?from=2025-06-15&to=2025-06-17", "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "2025-06-15", body["from"])
	require.Equal(t, "2025-06-17", body["to"])

	days := body["days"].([]any)
	require.Len(t, days, 3)
	require.Equal(t, "2025-06-15", days[0].(map[string]any)["date"])
	require.Len(t, days[0].(map[string]any)["planned"].([]any), 1)
	require.Len(t, days[1].(map[string]any)["planned"].([]any), 0, "June 16 empty")
	require.Len(t, days[2].(map[string]any)["actual"].([]any), 1)

	// to before from -> 400
	resp, _ = do(t, srv, http.MethodGet, "/api/timeline/range?from=2025-06-17&to=2025-06-15", "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// malformed date -> 400
	resp, _ = do(t, srv, http.MethodGet, "/api/timeline/range?from=not-a-date&to=2025-06-17", "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// missing params -> 400
	resp, _ = do(t, srv, http.MethodGet, "/api/timeline/range", "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestBlocksForTaskEndpoint(t *testing.T) {
	srv, pool, acc := httpSetupWithPool(t)

	var taskID uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO tasks (account_id, title, state) VALUES ($1, 'Write report', 'BACKLOG') RETURNING id`,
		acc).Scan(&taskID))

	resp, body := do(t, srv, http.MethodPost, "/api/blocks",
		`{"kind":"planned","date":"2025-06-15","start":"09:00","end":"10:00","task_id":"`+taskID.String()+`"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, taskID.String(), body["task_id"])

	resp, list := do(t, srv, http.MethodGet, "/api/tasks/"+taskID.String()+"/blocks", "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	blocks := list["blocks"].([]any)
	require.Len(t, blocks, 1)
	require.Equal(t, taskID.String(), blocks[0].(map[string]any)["task_id"])

	// unknown task -> 404
	resp, _ = do(t, srv, http.MethodGet, "/api/tasks/"+uuid.NewString()+"/blocks", "")
	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	// malformed id -> 404
	resp, _ = do(t, srv, http.MethodGet, "/api/tasks/not-a-uuid/blocks", "")
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestComparisonEndpoint_Range(t *testing.T) {
	srv, _ := httpSetup(t)

	do(t, srv, http.MethodPost, "/api/blocks",
		`{"kind":"actual","date":"2025-06-15","start":"09:00","end":"10:00"}`)
	do(t, srv, http.MethodPost, "/api/blocks",
		`{"kind":"actual","date":"2025-06-16","start":"09:00","end":"11:00"}`)

	resp, body := do(t, srv, http.MethodGet, "/api/comparison?from=2025-06-15&to=2025-06-16", "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "2025-06-15", body["from"])
	require.Equal(t, "2025-06-16", body["to"])
	require.Nil(t, body["date"], "range response carries no date field")

	cats := body["categories"].([]any)
	require.Len(t, cats, 1)
	require.Equal(t, float64(10800), cats[0].(map[string]any)["actual_seconds"])

	// to before from -> 400
	resp, _ = do(t, srv, http.MethodGet, "/api/comparison?from=2025-06-16&to=2025-06-15", "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// malformed from -> 400
	resp, _ = do(t, srv, http.MethodGet, "/api/comparison?from=not-a-date&to=2025-06-16", "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// plain ?date= still works and carries no from/to
	resp, body = do(t, srv, http.MethodGet, "/api/comparison?date=2025-06-15", "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "2025-06-15", body["date"])
	require.Nil(t, body["from"])
}
