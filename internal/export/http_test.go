package export_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/export"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
)

// stubProtector injects a fixed identity, standing in for the account middleware.
func stubProtector(accountID uuid.UUID) export.Protector {
	return func(fn http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn(w, r.WithContext(reqctx.WithIdentity(r.Context(), reqctx.Identity{AccountID: accountID})))
		})
	}
}

func do(t *testing.T, srv *httptest.Server) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/export", strings.NewReader(""))
	require.NoError(t, err)
	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	return resp
}

func TestExportEndpoint(t *testing.T) {
	e := setup(t)
	seed(t, e, e.acc)

	mux := http.NewServeMux()
	export.NewHandler(e.export).Mount(mux, stubProtector(e.acc))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	resp := do(t, srv)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))
	disposition := resp.Header.Get("Content-Disposition")
	require.Contains(t, disposition, "attachment")
	require.Contains(t, disposition, "productivity-os-export-")
	require.Contains(t, disposition, ".json")

	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	require.NotEmpty(t, body["exported_at"])
	require.Len(t, body["categories"], 2)
	require.Len(t, body["planned_blocks"], 1)
	require.Len(t, body["actual_blocks"], 1)
	require.Len(t, body["tasks"], 1)
	require.Len(t, body["habits"], 2)
	require.Len(t, body["habit_completions"], 2)
	require.Len(t, body["goals"], 1)
	require.Len(t, body["daily_reviews"], 1)
	require.Len(t, body["weekly_reviews"], 1)

	task := body["tasks"].([]any)[0].(map[string]any)
	require.Equal(t, "Ship export", task["title"])
}

func TestExportEndpoint_Isolation(t *testing.T) {
	e := setup(t)
	seed(t, e, e.acc)
	other := newAccount(t, e.pool, "other-http@test")

	mux := http.NewServeMux()
	export.NewHandler(e.export).Mount(mux, stubProtector(other))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	resp := do(t, srv)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Empty(t, body["categories"], "other sees none of the owner's export")
	require.Empty(t, body["tasks"])
	require.Empty(t, body["goals"])
}
