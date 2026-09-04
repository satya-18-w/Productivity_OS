package tasks_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
	"github.com/satya-18-w/productivity-os/internal/tasks"
)

func stubProtector(accountID uuid.UUID) tasks.Protector {
	return func(fn http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn(w, r.WithContext(reqctx.WithIdentity(r.Context(), reqctx.Identity{AccountID: accountID})))
		})
	}
}

func httpSetup(t *testing.T) *httptest.Server {
	t.Helper()
	pool := pgtest.Pool(t)
	acc := newAccount(t, pool, "http@test")
	mux := http.NewServeMux()
	tasks.NewHandler(tasks.NewService(pool)).Mount(mux, stubProtector(acc), stubProtector(acc))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
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
