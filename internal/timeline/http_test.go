package timeline_test

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
	pool := pgtest.Pool(t)
	acc := newAccount(t, pool, "http@test")
	mux := http.NewServeMux()
	z := fakeZone{}
	timeline.NewHandler(timeline.NewService(pool, z), z).Mount(mux, stubProtector(acc), stubProtector(acc))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, acc
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

func TestCategoryEndpoints(t *testing.T) {
	srv, _ := httpSetup(t)

	resp, body := do(t, srv, http.MethodPost, "/api/categories", `{"name":"Gym"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, "Gym", body["name"])
	id := body["id"].(string)

	// duplicate -> 409
	resp, body = do(t, srv, http.MethodPost, "/api/categories", `{"name":"gym"}`)
	require.Equal(t, http.StatusConflict, resp.StatusCode)
	require.Equal(t, "CONFLICT", body["error"].(map[string]any)["code"])

	// invalid -> 400
	resp, _ = do(t, srv, http.MethodPost, "/api/categories", `{"name":"  "}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// rename -> 204
	resp, _ = do(t, srv, http.MethodPatch, "/api/categories/"+id, `{"name":"Gym & Cardio"}`)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// list reflects it
	_, body = do(t, srv, http.MethodGet, "/api/categories", "")
	cats := body["categories"].([]any)
	require.Len(t, cats, 1)
	require.Equal(t, "Gym & Cardio", cats[0].(map[string]any)["name"])

	// archive -> 204, then gone from list
	resp, _ = do(t, srv, http.MethodPost, "/api/categories/"+id+"/archive", "")
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	_, body = do(t, srv, http.MethodGet, "/api/categories", "")
	require.Empty(t, body["categories"])

	// bad path id -> 404
	resp, _ = do(t, srv, http.MethodPatch, "/api/categories/not-a-uuid", `{"name":"X"}`)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
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
