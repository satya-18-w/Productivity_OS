package notes_test

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

	"github.com/satya-18-w/productivity-os/internal/notes"
	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
)

func setup(t *testing.T) (notes.Service, *pgxpool.Pool, uuid.UUID) {
	t.Helper()
	pool := pgtest.Pool(t)
	return notes.NewService(pool), pool, newAccount(t, pool, "owner@test")
}

func newAccount(t *testing.T, pool *pgxpool.Pool, email string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO accounts (email, password_hash, timezone) VALUES ($1, 'x', 'UTC') RETURNING id`,
		email).Scan(&id))
	return id
}

func TestNoteLifecycle(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	n, err := svc.CreateNote(ctx, acc, notes.NoteInput{Title: "  Idea  ", Body: "write it down"})
	require.NoError(t, err)
	require.Equal(t, "Idea", n.Title)
	require.Equal(t, "write it down", n.Body)

	require.NoError(t, svc.UpdateNote(ctx, acc, n.ID, notes.NoteInput{Title: "Better idea", Body: "revised"}))
	list, err := svc.ListNotes(ctx, acc)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "Better idea", list[0].Title)
	require.Equal(t, "revised", list[0].Body)

	require.NoError(t, svc.DeleteNote(ctx, acc, n.ID))
	require.ErrorIs(t, svc.DeleteNote(ctx, acc, n.ID), notes.ErrNoteNotFound)
	list, _ = svc.ListNotes(ctx, acc)
	require.Empty(t, list, "delete is permanent, no trash")
}

func TestNoteValidation(t *testing.T) {
	svc, _, acc := setup(t)
	ctx := context.Background()

	_, err := svc.CreateNote(ctx, acc, notes.NoteInput{Title: "  "})
	var verr *notes.ValidationError
	require.ErrorAs(t, err, &verr)
	require.Contains(t, verr.Fields, "title")

	_, err = svc.CreateNote(ctx, acc, notes.NoteInput{Title: strings.Repeat("x", 201)})
	require.ErrorAs(t, err, &verr)

	_, err = svc.CreateNote(ctx, acc, notes.NoteInput{Title: "ok", Body: strings.Repeat("y", 20001)})
	require.ErrorAs(t, err, &verr)
	require.Contains(t, verr.Fields, "body")
}

func TestNoteIsolation(t *testing.T) {
	svc, pool, a := setup(t)
	ctx := context.Background()
	b := newAccount(t, pool, "other@test")

	na, err := svc.CreateNote(ctx, a, notes.NoteInput{Title: "A's note"})
	require.NoError(t, err)
	_, err = svc.CreateNote(ctx, b, notes.NoteInput{Title: "B's note"})
	require.NoError(t, err)

	bl, err := svc.ListNotes(ctx, b)
	require.NoError(t, err)
	require.Len(t, bl, 1)
	require.Equal(t, "B's note", bl[0].Title)

	require.ErrorIs(t, svc.UpdateNote(ctx, b, na.ID, notes.NoteInput{Title: "x"}), notes.ErrNoteNotFound)
	require.ErrorIs(t, svc.DeleteNote(ctx, b, na.ID), notes.ErrNoteNotFound)
}

// --- HTTP ---

func stubProtector(accountID uuid.UUID) notes.Protector {
	return func(fn http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn(w, r.WithContext(reqctx.WithIdentity(r.Context(), reqctx.Identity{AccountID: accountID})))
		})
	}
}

func TestNoteEndpoints(t *testing.T) {
	pool := pgtest.Pool(t)
	acc := newAccount(t, pool, "http@test")
	mux := http.NewServeMux()
	notes.NewHandler(notes.NewService(pool)).Mount(mux, stubProtector(acc), stubProtector(acc))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	do := func(method, path, body string) (*http.Response, map[string]any) {
		req, _ := http.NewRequestWithContext(context.Background(), method, srv.URL+path, strings.NewReader(body))
		resp, err := srv.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		var m map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&m)
		return resp, m
	}

	resp, body := do(http.MethodPost, "/api/notes", `{"title":"Ship it","body":"MX4"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, "Ship it", body["title"])
	id := body["id"].(string)

	resp, _ = do(http.MethodPost, "/api/notes", `{"title":"  "}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp, _ = do(http.MethodPatch, "/api/notes/"+id, `{"title":"Edited","body":"still MX4"}`)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	_, list := do(http.MethodGet, "/api/notes", "")
	ns := list["notes"].([]any)
	require.Len(t, ns, 1)
	require.Equal(t, "Edited", ns[0].(map[string]any)["title"])
	require.Equal(t, "still MX4", ns[0].(map[string]any)["body"])

	resp, _ = do(http.MethodDelete, "/api/notes/"+id, "")
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp, _ = do(http.MethodDelete, "/api/notes/not-a-uuid", "")
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp, _ = do(http.MethodPatch, "/api/notes/"+uuid.NewString(), `{"title":"x"}`)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
