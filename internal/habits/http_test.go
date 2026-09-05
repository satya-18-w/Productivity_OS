package habits_test

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
	srv, _, _ := httpSetupWithPool(t)
	return srv
}

func httpSetupWithPool(t *testing.T) (*httptest.Server, *pgxpool.Pool, uuid.UUID) {
	t.Helper()
	pool := pgtest.Pool(t)
	acc := newAccount(t, pool, "http@test")
	mux := http.NewServeMux()
	habits.NewHandler(habits.NewService(pool, fakeZone{}, categories.NewService(pool)), fakeZone{}).
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

func TestHabitEndpoints_TargetAndEdit(t *testing.T) {
	srv := httpSetup(t)

	resp, body := do(t, srv, http.MethodPost, "/api/habits", `{"name":"Workout","target":"30 minutes"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, "30 minutes", body["target"])
	id := body["id"].(string)

	// full edit: name + target
	resp, body = do(t, srv, http.MethodPatch, "/api/habits/"+id, `{"name":"Workout daily","target":"45 minutes"}`)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "Workout daily", body["name"])
	require.Equal(t, "45 minutes", body["target"])

	// list reflects it
	_, list := do(t, srv, http.MethodGet, "/api/habits", "")
	h0 := list["habits"].([]any)[0].(map[string]any)
	require.Equal(t, "Workout daily", h0["name"])
	require.Equal(t, "45 minutes", h0["target"])

	// clearing the target (omit it)
	resp, body = do(t, srv, http.MethodPatch, "/api/habits/"+id, `{"name":"Workout daily"}`)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Nil(t, body["target"])

	// blank name -> 400
	resp, _ = do(t, srv, http.MethodPatch, "/api/habits/"+id, `{"name":"  "}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// unknown id -> 404
	resp, _ = do(t, srv, http.MethodPatch, "/api/habits/"+uuid.NewString(), `{"name":"X"}`)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
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

func TestHabitEndpoints_Category(t *testing.T) {
	srv, pool, acc := httpSetupWithPool(t)

	var catID uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO categories (account_id, name) VALUES ($1, 'Wellness') RETURNING id`, acc).Scan(&catID))

	resp, body := do(t, srv, http.MethodPost, "/api/habits", `{"name":"Meditate","category_id":"`+catID.String()+`"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, catID.String(), body["category_id"])
	id := body["id"].(string)

	// unknown category -> 400
	resp, _ = do(t, srv, http.MethodPost, "/api/habits",
		`{"name":"x","category_id":"`+uuid.NewString()+`"}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var other uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO categories (account_id, name) VALUES ($1, 'Focus') RETURNING id`, acc).Scan(&other))

	resp, _ = do(t, srv, http.MethodPut, "/api/habits/"+id+"/category", `{"category_id":"`+other.String()+`"}`)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	_, list := do(t, srv, http.MethodGet, "/api/habits", "")
	h0 := list["habits"].([]any)[0].(map[string]any)
	require.Equal(t, other.String(), h0["category_id"])

	// clear it
	resp, _ = do(t, srv, http.MethodPut, "/api/habits/"+id+"/category", `{}`)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	_, list = do(t, srv, http.MethodGet, "/api/habits", "")
	require.Nil(t, list["habits"].([]any)[0].(map[string]any)["category_id"])
}

func TestHabitEndpoint_Range(t *testing.T) {
	srv := httpSetup(t)
	_, body := do(t, srv, http.MethodPost, "/api/habits", `{"name":"Meditate"}`)
	id := body["id"].(string)

	d := today()
	yd := d.AddDays(-1).String()
	do(t, srv, http.MethodPut, "/api/habits/"+id+"/completions/"+d.String(), "")
	do(t, srv, http.MethodPut, "/api/habits/"+id+"/completions/"+yd, "")

	from := d.AddDays(-1).String()
	to := d.String()
	resp, list := do(t, srv, http.MethodGet, "/api/habits/range?from="+from+"&to="+to, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, from, list["from"])
	require.Equal(t, to, list["to"])
	hs := list["habits"].([]any)
	require.Len(t, hs, 1)
	require.Equal(t, float64(2), hs[0].(map[string]any)["count"])

	// to before from -> 400
	resp, _ = do(t, srv, http.MethodGet, "/api/habits/range?from="+to+"&to="+from, "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// malformed date -> 400
	resp, _ = do(t, srv, http.MethodGet, "/api/habits/range?from=nope&to="+to, "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHabitEndpoint_History(t *testing.T) {
	srv := httpSetup(t)
	_, body := do(t, srv, http.MethodPost, "/api/habits", `{"name":"Meditate"}`)
	id := body["id"].(string)

	d := today()
	yd := d.AddDays(-1).String()
	do(t, srv, http.MethodPut, "/api/habits/"+id+"/completions/"+d.String(), "")
	do(t, srv, http.MethodPut, "/api/habits/"+id+"/completions/"+yd, "")

	from := d.AddDays(-1).String()
	to := d.String()
	resp, list := do(t, srv, http.MethodGet, "/api/habits/history?from="+from+"&to="+to, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, from, list["from"])
	require.Equal(t, to, list["to"])
	hs := list["habits"].([]any)
	require.Len(t, hs, 1)
	h0 := hs[0].(map[string]any)
	require.Equal(t, false, h0["archived"])
	require.ElementsMatch(t, []any{from, to}, h0["completions"])

	// to before from -> 400
	resp, _ = do(t, srv, http.MethodGet, "/api/habits/history?from="+to+"&to="+from, "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// malformed date -> 400
	resp, _ = do(t, srv, http.MethodGet, "/api/habits/history?from=nope&to="+to, "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// range too large -> 400
	resp, _ = do(t, srv, http.MethodGet, "/api/habits/history?from="+d.AddDays(-93).String()+"&to="+to, "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHabitEndpoint_Week(t *testing.T) {
	srv := httpSetup(t)
	_, body := do(t, srv, http.MethodPost, "/api/habits", `{"name":"Workout"}`)
	id := body["id"].(string)

	d := today()
	do(t, srv, http.MethodPut, "/api/habits/"+id+"/completions/"+d.String(), "")

	resp, wk := do(t, srv, http.MethodGet, "/api/habits/week?date="+d.String(), "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	days := wk["days"].([]any)
	require.Len(t, days, 7)
	require.Contains(t, days, d.String())

	hs := wk["habits"].([]any)
	require.Len(t, hs, 1)
	h0 := hs[0].(map[string]any)
	require.Equal(t, "Workout", h0["name"])
	require.Equal(t, float64(1), h0["current_streak"])
	require.Contains(t, h0["completed"].([]any), d.String())

	require.Empty(t, wk["archived"])

	// missing date -> 400
	resp, _ = do(t, srv, http.MethodGet, "/api/habits/week", "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// malformed date -> 400
	resp, _ = do(t, srv, http.MethodGet, "/api/habits/week?date=nope", "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
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
