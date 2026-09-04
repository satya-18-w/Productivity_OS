package goals

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/platform/httpx"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

// Handler serves the goal HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler builds the goals handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Protector wraps a handler with auth (write also adds CSRF).
type Protector func(http.HandlerFunc) http.Handler

// Mount registers the goal routes.
func (h *Handler) Mount(mux *http.ServeMux, write, read Protector) {
	mux.Handle("GET /api/goals", read(h.list))
	mux.Handle("POST /api/goals", write(h.create))
	mux.Handle("PATCH /api/goals/{id}", write(h.update))
	mux.Handle("PUT /api/goals/{id}/progress", write(h.setProgress))
	mux.Handle("DELETE /api/goals/{id}", write(h.delete))
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

type goalBody struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	TargetDate  *string `json:"target_date"`
	Progress    string  `json:"progress"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func toBody(g Goal) goalBody {
	b := goalBody{
		ID:          g.ID.String(),
		Title:       g.Title,
		Description: g.Description,
		Progress:    string(g.Progress),
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
	if g.TargetDate != nil {
		s := g.TargetDate.String()
		b.TargetDate = &s
	}
	return b
}

type goalRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	TargetDate  *string `json:"target_date"`
}

func parseInput(req goalRequest) (GoalInput, *ValidationError) {
	in := GoalInput{Title: req.Title, Description: req.Description}
	if req.TargetDate != nil && *req.TargetDate != "" {
		d, err := timezone.ParseDate(*req.TargetDate)
		if err != nil {
			return GoalInput{}, &ValidationError{Fields: map[string]string{"target_date": "must be YYYY-MM-DD"}}
		}
		in.TargetDate = &d
	}
	return in, nil
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListGoals(r.Context(), accountID(r))
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	out := make([]goalBody, len(list))
	for i, g := range list {
		out[i] = toBody(g)
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"goals": out})
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req goalRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	in, verr := parseInput(req)
	if verr != nil {
		httpx.WriteError(w, r, httpx.ValidationError(verr.Fields))
		return
	}
	goal, err := h.svc.CreateGoal(r.Context(), accountID(r), in)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toBody(goal))
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var req goalRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	in, verr := parseInput(req)
	if verr != nil {
		httpx.WriteError(w, r, httpx.ValidationError(verr.Fields))
		return
	}
	if err := h.svc.UpdateGoal(r.Context(), accountID(r), id, in); err != nil {
		writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type progressRequest struct {
	Progress string `json:"progress"`
}

func (h *Handler) setProgress(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var req progressRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	if err := h.svc.SetProgress(r.Context(), accountID(r), id, Progress(req.Progress)); err != nil {
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
	if err := h.svc.DeleteGoal(r.Context(), accountID(r), id); err != nil {
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
	case errors.Is(err, ErrGoalNotFound):
		httpx.WriteError(w, r, httpx.NewError(http.StatusNotFound, httpx.CodeNotFound, "Goal not found"))
	default:
		httpx.WriteError(w, r, err)
	}
}
