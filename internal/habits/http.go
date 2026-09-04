package habits

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/platform/httpx"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

// Handler serves the habit HTTP endpoints.
type Handler struct {
	svc  Service
	zone AccountZone
}

// NewHandler builds the habits handler. zone resolves the account's "today" for
// the default habits view.
func NewHandler(svc Service, zone AccountZone) *Handler { return &Handler{svc: svc, zone: zone} }

// Protector wraps a handler with auth (write also adds CSRF).
type Protector func(http.HandlerFunc) http.Handler

// Mount registers the habit routes. write enforces auth + CSRF; read only auth.
func (h *Handler) Mount(mux *http.ServeMux, write, read Protector) {
	mux.Handle("GET /api/habits", read(h.list))
	mux.Handle("POST /api/habits", write(h.create))
	mux.Handle("POST /api/habits/{id}/archive", write(h.archive))
	mux.Handle("POST /api/habits/{id}/unarchive", write(h.unarchive))
	mux.Handle("PUT /api/habits/{id}/completions/{date}", write(h.mark))
	mux.Handle("DELETE /api/habits/{id}/completions/{date}", write(h.unmark))
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

func pathDate(w http.ResponseWriter, r *http.Request) (timezone.Date, bool) {
	d, err := timezone.ParseDate(r.PathValue("date"))
	if err != nil {
		httpx.WriteError(w, r, httpx.ValidationError(map[string]string{"date": "must be YYYY-MM-DD"}))
		return timezone.Date{}, false
	}
	return d, true
}

type habitBody struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type habitViewBody struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	CurrentStreak   int    `json:"current_streak"`
	CompletedOnDate bool   `json:"completed_on_date"`
	Last30Days      int    `json:"last_30_days"`
}

type listResponse struct {
	Date     string          `json:"date"`
	Habits   []habitViewBody `json:"habits"`
	Archived []habitBody     `json:"archived"`
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	acc := accountID(r)

	var viewDate timezone.Date
	if raw := r.URL.Query().Get("date"); raw != "" {
		d, err := timezone.ParseDate(raw)
		if err != nil {
			httpx.WriteError(w, r, httpx.ValidationError(map[string]string{"date": "must be YYYY-MM-DD"}))
			return
		}
		viewDate = d
	} else {
		loc, err := h.zone.Zone(r.Context(), acc)
		if err != nil {
			httpx.WriteError(w, r, err)
			return
		}
		viewDate = timezone.Today(loc)
	}

	views, err := h.svc.ListActive(r.Context(), acc, viewDate)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	archived, err := h.svc.ListArchived(r.Context(), acc)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	resp := listResponse{Date: viewDate.String(), Habits: []habitViewBody{}, Archived: []habitBody{}}
	for _, v := range views {
		resp.Habits = append(resp.Habits, habitViewBody{
			ID: v.ID.String(), Name: v.Name,
			CurrentStreak: v.CurrentStreak, CompletedOnDate: v.CompletedOnDate, Last30Days: v.Last30Days,
		})
	}
	for _, a := range archived {
		resp.Archived = append(resp.Archived, habitBody{ID: a.ID.String(), Name: a.Name})
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

type nameRequest struct {
	Name string `json:"name"`
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req nameRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	habit, err := h.svc.CreateHabit(r.Context(), accountID(r), req.Name)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, habitBody{ID: habit.ID.String(), Name: habit.Name})
}

func (h *Handler) archive(w http.ResponseWriter, r *http.Request)   { h.setArchived(w, r, true) }
func (h *Handler) unarchive(w http.ResponseWriter, r *http.Request) { h.setArchived(w, r, false) }

func (h *Handler) setArchived(w http.ResponseWriter, r *http.Request, archive bool) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var err error
	if archive {
		err = h.svc.ArchiveHabit(r.Context(), accountID(r), id)
	} else {
		err = h.svc.UnarchiveHabit(r.Context(), accountID(r), id)
	}
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) mark(w http.ResponseWriter, r *http.Request)   { h.setCompletion(w, r, true) }
func (h *Handler) unmark(w http.ResponseWriter, r *http.Request) { h.setCompletion(w, r, false) }

func (h *Handler) setCompletion(w http.ResponseWriter, r *http.Request, mark bool) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	date, ok := pathDate(w, r)
	if !ok {
		return
	}
	var err error
	if mark {
		err = h.svc.MarkComplete(r.Context(), accountID(r), id, date)
	} else {
		err = h.svc.UnmarkComplete(r.Context(), accountID(r), id, date)
	}
	if err != nil {
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
	case errors.Is(err, ErrHabitNotFound):
		httpx.WriteError(w, r, httpx.NewError(http.StatusNotFound, httpx.CodeNotFound, "Habit not found"))
	default:
		httpx.WriteError(w, r, err)
	}
}
