package categories_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/categories"
	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
)

// stubProtector injects a fixed identity, standing in for the account middleware.
func stubProtector(accountID uuid.UUID) categories.Protector {
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
	categories.NewHandler(categories.NewService(pool)).Mount(mux, stubProtector(acc), stubProtector(acc))
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

func TestCategoryEndpoints(t *testing.T) {
	srv := httpSetup(t)

	resp, body := do(t, srv, http.MethodPost, "/api/categories",
		`{"name":"Gym","colour":"green","icon":"dumbbell"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, "Gym", body["name"])
	require.Equal(t, "green", body["colour"])
	require.Equal(t, "dumbbell", body["icon"])
	id := body["id"].(string)

	// duplicate -> 409
	resp, body = do(t, srv, http.MethodPost, "/api/categories", `{"name":"gym"}`)
	require.Equal(t, http.StatusConflict, resp.StatusCode)
	require.Equal(t, "CONFLICT", body["error"].(map[string]any)["code"])

	// invalid -> 400
	resp, _ = do(t, srv, http.MethodPost, "/api/categories", `{"name":"  "}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// update name + colour -> 204
	resp, _ = do(t, srv, http.MethodPatch, "/api/categories/"+id,
		`{"name":"Gym & Cardio","colour":"orange","icon":"dumbbell"}`)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// list reflects it
	_, body = do(t, srv, http.MethodGet, "/api/categories", "")
	cats := body["categories"].([]any)
	require.Len(t, cats, 1)
	c0 := cats[0].(map[string]any)
	require.Equal(t, "Gym & Cardio", c0["name"])
	require.Equal(t, "orange", c0["colour"])

	// archive -> 204, then gone from list
	resp, _ = do(t, srv, http.MethodPost, "/api/categories/"+id+"/archive", "")
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	_, body = do(t, srv, http.MethodGet, "/api/categories", "")
	require.Empty(t, body["categories"])

	// bad path id -> 404
	resp, _ = do(t, srv, http.MethodPatch, "/api/categories/not-a-uuid", `{"name":"X"}`)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestCategoryEndpoints_PartialUpdate is CP R3: a PATCH with only {"name":"..."}
// must leave colour/icon untouched (docs/left.md's Category C.1 gap).
func TestCategoryEndpoints_PartialUpdate(t *testing.T) {
	srv := httpSetup(t)

	_, body := do(t, srv, http.MethodPost, "/api/categories", `{"name":"Gym","colour":"blue","icon":"dumbbell"}`)
	id := body["id"].(string)

	// rename-only PATCH — colour/icon omitted, must survive
	resp, _ := do(t, srv, http.MethodPatch, "/api/categories/"+id, `{"name":"Gym & Cardio"}`)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	_, body = do(t, srv, http.MethodGet, "/api/categories", "")
	c0 := body["categories"].([]any)[0].(map[string]any)
	require.Equal(t, "Gym & Cardio", c0["name"])
	require.Equal(t, "blue", c0["colour"], "omitted colour is untouched by a rename-only PATCH")
	require.Equal(t, "dumbbell", c0["icon"], "omitted icon is untouched by a rename-only PATCH")

	// colour-only PATCH — name/icon untouched
	resp, _ = do(t, srv, http.MethodPatch, "/api/categories/"+id, `{"colour":"orange"}`)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	_, body = do(t, srv, http.MethodGet, "/api/categories", "")
	c0 = body["categories"].([]any)[0].(map[string]any)
	require.Equal(t, "Gym & Cardio", c0["name"])
	require.Equal(t, "orange", c0["colour"])
	require.Equal(t, "dumbbell", c0["icon"])

	// explicit empty icon clears it
	resp, _ = do(t, srv, http.MethodPatch, "/api/categories/"+id, `{"icon":""}`)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	_, body = do(t, srv, http.MethodGet, "/api/categories", "")
	c0 = body["categories"].([]any)[0].(map[string]any)
	require.Equal(t, "orange", c0["colour"])
	require.Equal(t, "", c0["icon"])
}
