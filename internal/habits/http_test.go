package habits_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/habits"
	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
)

func stubProtector(accountID uuid.UUID) habits.Protector {
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
	habits.NewHandler(habits.NewService(pool, fakeZone{}), fakeZone{}).
		Mount(mux, stubProtector(acc), stubProtector(acc))
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

func TestHabitEndpoints(t *testing.T) {
	srv := httpSetup(t)
	d := today().String()
	yd := today().AddDays(-1).String()

	// create
	resp, body := do(t, srv, http.MethodPost, "/api/habits", `{"name":"Read"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	id := body["id"].(string)

	// blank name -> 400
	resp, _ = do(t, srv, http.MethodPost, "/api/habits", `{"name":"  "}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// mark today + yesterday (mark is idempotent)
	for _, date := range []string{d, d, yd} {
		resp, _ = do(t, srv, http.MethodPut, "/api/habits/"+id+"/completions/"+date, "")
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
	}
	// bad date -> 400
	resp, _ = do(t, srv, http.MethodPut, "/api/habits/"+id+"/completions/2025-99-99", "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// list carries the streak
	_, list := do(t, srv, http.MethodGet, "/api/habits", "")
	hs := list["habits"].([]any)
	require.Len(t, hs, 1)
	h0 := hs[0].(map[string]any)
	require.Equal(t, float64(2), h0["current_streak"])
	require.Equal(t, true, h0["completed_on_date"])

	// unmark yesterday -> streak drops to 1
	resp, _ = do(t, srv, http.MethodDelete, "/api/habits/"+id+"/completions/"+yd, "")
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	_, list = do(t, srv, http.MethodGet, "/api/habits", "")
	require.Equal(t, float64(1), list["habits"].([]any)[0].(map[string]any)["current_streak"])

	// archive -> gone from habits, present in archived
	resp, _ = do(t, srv, http.MethodPost, "/api/habits/"+id+"/archive", "")
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	_, list = do(t, srv, http.MethodGet, "/api/habits", "")
	require.Empty(t, list["habits"])
	require.Len(t, list["archived"].([]any), 1)

	// unarchive -> back
	resp, _ = do(t, srv, http.MethodPost, "/api/habits/"+id+"/unarchive", "")
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	_, list = do(t, srv, http.MethodGet, "/api/habits", "")
	require.Len(t, list["habits"].([]any), 1)

	// unknown / malformed id
	resp, _ = do(t, srv, http.MethodPost, "/api/habits/"+uuid.NewString()+"/archive", "")
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp, _ = do(t, srv, http.MethodPut, "/api/habits/not-a-uuid/completions/"+d, "")
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHabitEndpoint_ViewDate(t *testing.T) {
	srv := httpSetup(t)
	_, body := do(t, srv, http.MethodPost, "/api/habits", `{"name":"Stretch"}`)
	id := body["id"].(string)

	past := today().AddDays(-5).String()
	do(t, srv, http.MethodPut, "/api/habits/"+id+"/completions/"+past, "")

	_, list := do(t, srv, http.MethodGet, "/api/habits?date="+past, "")
	require.Equal(t, true, list["habits"].([]any)[0].(map[string]any)["completed_on_date"])

	_, list = do(t, srv, http.MethodGet, "/api/habits", "") // today
	require.Equal(t, false, list["habits"].([]any)[0].(map[string]any)["completed_on_date"])
}
