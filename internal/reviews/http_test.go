package reviews_test

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
	"github.com/satya-18-w/productivity-os/internal/reviews"
)

// stubProtector injects a fixed identity, standing in for the account middleware.
func stubProtector(accountID uuid.UUID) reviews.Protector {
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
	reviews.NewHandler(reviews.NewService(pool)).Mount(mux, stubProtector(acc), stubProtector(acc))
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

func TestDailyReviewEndpoints(t *testing.T) {
	srv := httpSetup(t)

	// blank before any save: prompts present, answers empty
	resp, body := do(t, srv, http.MethodGet, "/api/reviews/daily?date=2025-06-15", "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "2025-06-15", body["date"])
	prompts := body["prompts"].([]any)
	require.Len(t, prompts, 4)
	require.Empty(t, body["answers"])
	require.Nil(t, body["updated_at"])

	// save
	resp, _ = do(t, srv, http.MethodPut, "/api/reviews/daily?date=2025-06-15",
		`{"answers":{"went_well":"Shipped M6","not_a_real_prompt":"dropped"}}`)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// reload reflects it, unknown key dropped
	_, body = do(t, srv, http.MethodGet, "/api/reviews/daily?date=2025-06-15", "")
	answers := body["answers"].(map[string]any)
	require.Equal(t, "Shipped M6", answers["went_well"])
	require.NotContains(t, answers, "not_a_real_prompt")
	require.NotEmpty(t, body["updated_at"])

	// missing date -> 400
	resp, _ = do(t, srv, http.MethodGet, "/api/reviews/daily", "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp, _ = do(t, srv, http.MethodPut, "/api/reviews/daily", `{"answers":{}}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// bad date -> 400
	resp, _ = do(t, srv, http.MethodGet, "/api/reviews/daily?date=not-a-date", "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestWeeklyReviewEndpoints(t *testing.T) {
	srv := httpSetup(t)

	// blank before any save
	resp, body := do(t, srv, http.MethodGet, "/api/reviews/weekly?year=2025&week=24", "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, float64(2025), body["iso_year"])
	require.Equal(t, float64(24), body["iso_week"])
	require.Len(t, body["prompts"].([]any), 4)
	require.Empty(t, body["answers"])

	// save
	resp, _ = do(t, srv, http.MethodPut, "/api/reviews/weekly?year=2025&week=24",
		`{"answers":{"highlights":"Shipped M6"}}`)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	_, body = do(t, srv, http.MethodGet, "/api/reviews/weekly?year=2025&week=24", "")
	answers := body["answers"].(map[string]any)
	require.Equal(t, "Shipped M6", answers["highlights"])

	// missing year/week -> 400
	resp, _ = do(t, srv, http.MethodGet, "/api/reviews/weekly?year=2025", "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp, _ = do(t, srv, http.MethodGet, "/api/reviews/weekly?year=nope&week=24", "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// out-of-range week -> 400
	resp, _ = do(t, srv, http.MethodGet, "/api/reviews/weekly?year=2025&week=54", "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp, _ = do(t, srv, http.MethodPut, "/api/reviews/weekly?year=2025&week=0", `{"answers":{}}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestReviewEndpoints_Isolation(t *testing.T) {
	pool := pgtest.Pool(t)
	a := newAccount(t, pool, "a@test")
	b := newAccount(t, pool, "b@test")

	mux := http.NewServeMux()
	svc := reviews.NewService(pool)
	reviews.NewHandler(svc).Mount(mux, stubProtector(a), stubProtector(a))
	srvA := httptest.NewServer(mux)
	t.Cleanup(srvA.Close)

	muxB := http.NewServeMux()
	reviews.NewHandler(svc).Mount(muxB, stubProtector(b), stubProtector(b))
	srvB := httptest.NewServer(muxB)
	t.Cleanup(srvB.Close)

	resp, _ := do(t, srvA, http.MethodPut, "/api/reviews/daily?date=2025-06-15",
		`{"answers":{"went_well":"A's answer"}}`)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	_, body := do(t, srvB, http.MethodGet, "/api/reviews/daily?date=2025-06-15", "")
	require.Empty(t, body["answers"], "B sees none of A's review")
}
