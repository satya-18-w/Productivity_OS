package notes

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/platform/httpx"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
)

// Handler serves the note HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler builds the notes handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Protector wraps a handler with auth (write also adds CSRF).
type Protector func(http.HandlerFunc) http.Handler

// Mount registers the note routes. write enforces auth + CSRF; read only auth.
func (h *Handler) Mount(mux *http.ServeMux, write, read Protector) {
	mux.Handle("GET /api/notes", read(h.list))
	mux.Handle("POST /api/notes", write(h.create))
	mux.Handle("PATCH /api/notes/{id}", write(h.update))
	mux.Handle("DELETE /api/notes/{id}", write(h.delete))
}

func accountID(r *http.Request) uuid.UUID {
	id, _ := reqctx.IdentityFrom(r.Context())
	return id.AccountID
}

func pathID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, r, httpx.NewError(http.StatusNotFound, httpx.CodeNotFound, "Not found"))
		return uuid.Nil, false
	}
	return id, true
}

type noteBody struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toBody(n Note) noteBody {
	return noteBody{
		ID:        n.ID.String(),
		Title:     n.Title,
		Body:      n.Body,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}

type noteRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func parseInput(req noteRequest) NoteInput {
	return NoteInput(req)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListNotes(r.Context(), accountID(r))
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	out := make([]noteBody, len(list))
	for i, n := range list {
		out[i] = toBody(n)
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"notes": out})
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req noteRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	note, err := h.svc.CreateNote(r.Context(), accountID(r), parseInput(req))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toBody(note))
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var req noteRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	if err := h.svc.UpdateNote(r.Context(), accountID(r), id, parseInput(req)); err != nil {
		writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteNote(r.Context(), accountID(r), id); err != nil {
		writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var verr *ValidationError
	switch {
	case errors.As(err, &verr):
		httpx.WriteError(w, r, httpx.ValidationError(verr.Fields))
	case errors.Is(err, ErrNoteNotFound):
		httpx.WriteError(w, r, httpx.NewError(http.StatusNotFound, httpx.CodeNotFound, "Note not found"))
	default:
		httpx.WriteError(w, r, err)
	}
}
