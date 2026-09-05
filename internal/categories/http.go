package categories

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/platform/httpx"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
)

// Handler serves the categories HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler builds the categories handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Protector wraps a handler with auth (write also adds CSRF). cmd/server supplies
// the account module's middleware.
type Protector func(http.HandlerFunc) http.Handler

// Mount registers the categories routes. write must enforce auth + CSRF; read only
// auth.
func (h *Handler) Mount(mux *http.ServeMux, write, read Protector) {
	mux.Handle("GET /api/categories", read(h.list))
	mux.Handle("POST /api/categories", write(h.create))
	mux.Handle("PATCH /api/categories/{id}", write(h.update))
	mux.Handle("POST /api/categories/{id}/archive", write(h.archive))
}

type categoryBody struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Colour string `json:"colour"`
	Icon   string `json:"icon"`
}

func toBody(c Category) categoryBody {
	return categoryBody{ID: c.ID.String(), Name: c.Name, Colour: c.Colour, Icon: c.Icon}
}

// categoryRequest is the full category representation for create. name is
// required; colour and icon are optional keys (empty leaves them unset).
type categoryRequest struct {
	Name   string `json:"name"`
	Colour string `json:"colour"`
	Icon   string `json:"icon"`
}

func (r categoryRequest) input() Input {
	return Input(r)
}

// categoryUpdateRequest is a partial update (R3) — an omitted JSON key decodes to
// a nil pointer and leaves that field unchanged; a present key (including "")
// decodes to a non-nil pointer and sets it.
type categoryUpdateRequest struct {
	Name   *string `json:"name"`
	Colour *string `json:"colour"`
	Icon   *string `json:"icon"`
}

func (r categoryUpdateRequest) input() UpdateInput {
	return UpdateInput(r)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	cats, err := h.svc.List(r.Context(), accountID(r))
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	out := make([]categoryBody, len(cats))
	for i, c := range cats {
		out[i] = toBody(c)
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"categories": out})
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req categoryRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	cat, err := h.svc.Create(r.Context(), accountID(r), req.input())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toBody(cat))
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var req categoryUpdateRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	if _, err := h.svc.Update(r.Context(), accountID(r), id, req.input()); err != nil {
		writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) archive(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.svc.Archive(r.Context(), accountID(r), id); err != nil {
		writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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

func writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var verr *ValidationError
	switch {
	case errors.As(err, &verr):
		httpx.WriteError(w, r, httpx.ValidationError(verr.Fields))
	case errors.Is(err, ErrNameTaken):
		httpx.WriteError(w, r, httpx.NewError(http.StatusConflict, httpx.CodeConflict,
			"A category with this name already exists"))
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.NewError(http.StatusNotFound, httpx.CodeNotFound, "Category not found"))
	default:
		httpx.WriteError(w, r, err)
	}
}
