package tasks

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/platform/httpx"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

// Handler serves the task and board HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler builds the tasks handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Protector wraps a handler with auth (write also adds CSRF). cmd/server supplies
// the account module's middleware.
type Protector func(http.HandlerFunc) http.Handler

// Mount registers the task routes. write enforces auth + CSRF; read only auth.
func (h *Handler) Mount(mux *http.ServeMux, write, read Protector) {
	mux.Handle("GET /api/board", read(h.getBoard))
	mux.Handle("POST /api/tasks", write(h.createTask))
	mux.Handle("PATCH /api/tasks/{id}", write(h.updateTask))
	mux.Handle("PUT /api/tasks/{id}/state", write(h.moveTask))
	mux.Handle("DELETE /api/tasks/{id}", write(h.deleteTask))
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

type taskBody struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	DueDate     *string `json:"due_date"`
	State       string  `json:"state"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func toTaskBody(t Task) taskBody {
	b := taskBody{
		ID:          t.ID.String(),
		Title:       t.Title,
		Description: t.Description,
		State:       string(t.State),
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
	if t.DueDate != nil {
		s := t.DueDate.String()
		b.DueDate = &s
	}
	return b
}

type boardColumnBody struct {
	State string     `json:"state"`
	Tasks []taskBody `json:"tasks"`
}

type boardBody struct {
	Columns []boardColumnBody `json:"columns"`
}

func (h *Handler) getBoard(w http.ResponseWriter, r *http.Request) {
	board, err := h.svc.Board(r.Context(), accountID(r))
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	out := boardBody{Columns: make([]boardColumnBody, len(board.Columns))}
	for i, col := range board.Columns {
		tb := make([]taskBody, len(col.Tasks))
		for j, t := range col.Tasks {
			tb[j] = toTaskBody(t)
		}
		out.Columns[i] = boardColumnBody{State: string(col.State), Tasks: tb}
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

type taskRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	DueDate     *string `json:"due_date"`
}

func parseTaskInput(req taskRequest) (TaskInput, *ValidationError) {
	in := TaskInput{Title: req.Title, Description: req.Description}
	if req.DueDate != nil && *req.DueDate != "" {
		d, err := timezone.ParseDate(*req.DueDate)
		if err != nil {
			return TaskInput{}, &ValidationError{Fields: map[string]string{"due_date": "must be YYYY-MM-DD"}}
		}
		in.DueDate = &d
	}
	return in, nil
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	var req taskRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	in, verr := parseTaskInput(req)
	if verr != nil {
		httpx.WriteError(w, r, httpx.ValidationError(verr.Fields))
		return
	}
	task, err := h.svc.CreateTask(r.Context(), accountID(r), in)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toTaskBody(task))
}

func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var req taskRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	in, verr := parseTaskInput(req)
	if verr != nil {
		httpx.WriteError(w, r, httpx.ValidationError(verr.Fields))
		return
	}
	if err := h.svc.UpdateTask(r.Context(), accountID(r), id, in); err != nil {
		writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type stateRequest struct {
	State string `json:"state"`
}

func (h *Handler) moveTask(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var req stateRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	if err := h.svc.MoveTask(r.Context(), accountID(r), id, State(req.State)); err != nil {
		writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteTask(r.Context(), accountID(r), id); err != nil {
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
	case errors.Is(err, ErrTaskNotFound):
		httpx.WriteError(w, r, httpx.NewError(http.StatusNotFound, httpx.CodeNotFound, "Task not found"))
	default:
		httpx.WriteError(w, r, err)
	}
}
