package tasks_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/categories"
	"github.com/satya-18-w/productivity-os/internal/goals"
	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
	"github.com/satya-18-w/productivity-os/internal/tasks"
)

type fakeZone struct{}

func (fakeZone) Zone(context.Context, uuid.UUID) (*time.Location, error) { return time.UTC, nil }

func stubProtector(accountID uuid.UUID) tasks.Protector {
	return func(fn http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn(w, r.WithContext(reqctx.WithIdentity(r.Context(), reqctx.Identity{AccountID: accountID})))
		})
	}
}

func httpSetup(t *testing.T) *httptest.Server {
	srv, _, _ := httpSetupWithPool(t)
	return srv
}

func httpSetupWithPool(t *testing.T) (*httptest.Server, *pgxpool.Pool, uuid.UUID) {
	t.Helper()
	pool := pgtest.Pool(t)
	acc := newAccount(t, pool, "http@test")
	mux := http.NewServeMux()
	catSvc := categories.NewService(pool)
	tasks.NewHandler(tasks.NewService(pool, catSvc, goals.NewService(pool, catSvc)), fakeZone{}).
		Mount(mux, stubProtector(acc), stubProtector(acc))
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

func TestTaskEndpoints(t *testing.T) {
	srv := httpSetup(t)

	// create
	resp, body := do(t, srv, http.MethodPost, "/api/tasks",
		`{"title":"Write tests","description":"cover the board","due_date":"2025-07-04"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, "BACKLOG", body["state"])
	require.Equal(t, "2025-07-04", body["due_date"])
	id := body["id"].(string)

	// validation
	resp, _ = do(t, srv, http.MethodPost, "/api/tasks", `{"title":"  "}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp, _ = do(t, srv, http.MethodPost, "/api/tasks", `{"title":"x","due_date":"04-07-2025"}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// update
	resp, _ = do(t, srv, http.MethodPatch, "/api/tasks/"+id, `{"title":"Write more tests","description":""}`)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// move through all four columns
	for _, st := range []string{"TODO", "IN_PROGRESS", "DONE", "BACKLOG"} {
		resp, _ = do(t, srv, http.MethodPut, "/api/tasks/"+id+"/state", `{"state":"`+st+`"}`)
		require.Equalf(t, http.StatusNoContent, resp.StatusCode, "move to %s", st)
	}

	// bad state
	resp, _ = do(t, srv, http.MethodPut, "/api/tasks/"+id+"/state", `{"state":"NOWHERE"}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// unknown id
	resp, _ = do(t, srv, http.MethodPut, "/api/tasks/"+uuid.NewString()+"/state", `{"state":"DONE"}`)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp, _ = do(t, srv, http.MethodPatch, "/api/tasks/not-a-uuid", `{"title":"x"}`)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	// delete
	resp, _ = do(t, srv, http.MethodDelete, "/api/tasks/"+id, "")
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestTaskEndpoints_Category(t *testing.T) {
	srv, pool, acc := httpSetupWithPool(t)

	var catID uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO categories (account_id, name) VALUES ($1, 'Deep Work') RETURNING id`, acc).Scan(&catID))

	resp, body := do(t, srv, http.MethodPost, "/api/tasks", `{"title":"Ship it","category_id":"`+catID.String()+`"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, catID.String(), body["category_id"])

	resp, _ = do(t, srv, http.MethodPost, "/api/tasks",
		`{"title":"x","category_id":"`+uuid.NewString()+`"}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, "unknown category -> 400")

	resp, _ = do(t, srv, http.MethodPost, "/api/tasks", `{"title":"x","category_id":"not-a-uuid"}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestTaskEndpoints_Priority(t *testing.T) {
	srv := httpSetup(t)

	resp, body := do(t, srv, http.MethodPost, "/api/tasks", `{"title":"Ship it","priority":"HIGH"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, "HIGH", body["priority"])
	id := body["id"].(string)

	resp, _ = do(t, srv, http.MethodPatch, "/api/tasks/"+id, `{"title":"Ship it","priority":"LOW"}`)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	_, board := do(t, srv, http.MethodGet, "/api/board", "")
	cols := board["columns"].([]any)
	backlog := cols[0].(map[string]any)["tasks"].([]any)
	require.Equal(t, "LOW", backlog[0].(map[string]any)["priority"])

	resp, _ = do(t, srv, http.MethodPost, "/api/tasks", `{"title":"x","priority":"URGENT"}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, "invalid priority -> 400")
}

func TestThroughputEndpoint(t *testing.T) {
	srv := httpSetup(t) // fakeZone => UTC

	_, t1 := do(t, srv, http.MethodPost, "/api/tasks", `{"title":"one"}`)
	_, t2 := do(t, srv, http.MethodPost, "/api/tasks", `{"title":"two"}`)
	do(t, srv, http.MethodPut, "/api/tasks/"+t1["id"].(string)+"/state", `{"state":"DONE"}`)
	do(t, srv, http.MethodPut, "/api/tasks/"+t2["id"].(string)+"/state", `{"state":"TODO"}`)
	do(t, srv, http.MethodPut, "/api/tasks/"+t2["id"].(string)+"/state", `{"state":"DONE"}`)
	// bounce t1 out and back into DONE — still counts once
	do(t, srv, http.MethodPut, "/api/tasks/"+t1["id"].(string)+"/state", `{"state":"TODO"}`)
	do(t, srv, http.MethodPut, "/api/tasks/"+t1["id"].(string)+"/state", `{"state":"DONE"}`)

	today := time.Now().UTC().Format("2006-01-02")
	resp, body := do(t, srv, http.MethodGet, "/api/tasks/throughput?from="+today+"&to="+today, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, float64(2), body["done_count"])
	require.Equal(t, today, body["from"])
	require.Equal(t, today, body["to"])

	// to before from -> 400
	resp, _ = do(t, srv, http.MethodGet, "/api/tasks/throughput?from="+today+"&to=2000-01-01", "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// malformed date -> 400
	resp, _ = do(t, srv, http.MethodGet, "/api/tasks/throughput?from=nope&to="+today, "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestBoardEndpoint(t *testing.T) {
	srv := httpSetup(t)

	c, first := do(t, srv, http.MethodPost, "/api/tasks", `{"title":"one"}`)
	require.Equal(t, http.StatusCreated, c.StatusCode)
	do(t, srv, http.MethodPost, "/api/tasks", `{"title":"two"}`)
	do(t, srv, http.MethodPut, "/api/tasks/"+first["id"].(string)+"/state", `{"state":"DONE"}`)

	_, board := do(t, srv, http.MethodGet, "/api/board", "")
	cols := board["columns"].([]any)
	require.Len(t, cols, 4)

	states := make([]string, 4)
	for i, c := range cols {
		states[i] = c.(map[string]any)["state"].(string)
	}
	require.Equal(t, []string{"BACKLOG", "TODO", "IN_PROGRESS", "DONE"}, states)

	require.Len(t, cols[0].(map[string]any)["tasks"].([]any), 1) // "two"
	require.Len(t, cols[3].(map[string]any)["tasks"].([]any), 1) // "one" -> DONE
}
