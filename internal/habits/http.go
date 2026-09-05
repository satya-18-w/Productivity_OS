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
	mux.Handle("GET /api/habits/range", read(h.rangeCounts))
	mux.Handle("GET /api/habits/history", read(h.history))
	mux.Handle("GET /api/habits/week", read(h.week))
	mux.Handle("POST /api/habits", write(h.create))
	mux.Handle("PATCH /api/habits/{id}", write(h.update))
	mux.Handle("PUT /api/habits/{id}/category", write(h.setCategory))
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
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	CategoryID *string `json:"category_id"`
	Target     *string `json:"target"`
}

func toHabitBody(h Habit) habitBody {
	return habitBody{ID: h.ID.String(), Name: h.Name, CategoryID: categoryIDString(h.CategoryID), Target: h.Target}
}

func categoryIDString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

type habitViewBody struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	CategoryID      *string `json:"category_id"`
	Target          *string `json:"target"`
	CurrentStreak   int     `json:"current_streak"`
	CompletedOnDate bool    `json:"completed_on_date"`
	Last30Days      int     `json:"last_30_days"`
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
			ID: v.ID.String(), Name: v.Name, CategoryID: categoryIDString(v.CategoryID), Target: v.Target,
			CurrentStreak: v.CurrentStreak, CompletedOnDate: v.CompletedOnDate, Last30Days: v.Last30Days,
		})
	}
	for _, a := range archived {
		resp.Archived = append(resp.Archived, toHabitBody(a))
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

type rangeCountBody struct {
	HabitID string `json:"habit_id"`
	Name    string `json:"name"`
	Count   int    `json:"count"`
}

type rangeResponse struct {
	From   string           `json:"from"`
	To     string           `json:"to"`
	Habits []rangeCountBody `json:"habits"`
}

// rangeCounts serves GET /api/habits/range?from=&to= — every habit's completion
// count over the inclusive date range (M6/M7 foundation).
func (h *Handler) rangeCounts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fields := map[string]string{}

	from, fromErr := timezone.ParseDate(q.Get("from"))
	if fromErr != nil {
		fields["from"] = "must be YYYY-MM-DD"
	}
	to, toErr := timezone.ParseDate(q.Get("to"))
	if toErr != nil {
		fields["to"] = "must be YYYY-MM-DD"
	}
	if len(fields) > 0 {
		httpx.WriteError(w, r, httpx.ValidationError(fields))
		return
	}

	counts, err := h.svc.CompletionCountsInRange(r.Context(), accountID(r), from, to)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	resp := rangeResponse{From: from.String(), To: to.String(), Habits: []rangeCountBody{}}
	for _, c := range counts {
		resp.Habits = append(resp.Habits, rangeCountBody{HabitID: c.HabitID.String(), Name: c.Name, Count: c.Count})
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

type habitHistoryEntryBody struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Archived    bool     `json:"archived"`
	Completions []string `json:"completions"`
}

type historyResponse struct {
	From   string                  `json:"from"`
	To     string                  `json:"to"`
	Habits []habitHistoryEntryBody `json:"habits"`
}

// history serves GET /api/habits/history?from=&to= — every habit's completion
// dates within the inclusive range, for the "This Month" heatmap (R2,
// docs/left.md Phase 6).
func (h *Handler) history(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fields := map[string]string{}

	from, fromErr := timezone.ParseDate(q.Get("from"))
	if fromErr != nil {
		fields["from"] = "must be YYYY-MM-DD"
	}
	to, toErr := timezone.ParseDate(q.Get("to"))
	if toErr != nil {
		fields["to"] = "must be YYYY-MM-DD"
	}
	if len(fields) > 0 {
		httpx.WriteError(w, r, httpx.ValidationError(fields))
		return
	}

	entries, err := h.svc.History(r.Context(), accountID(r), from, to)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	resp := historyResponse{From: from.String(), To: to.String(), Habits: []habitHistoryEntryBody{}}
	for _, e := range entries {
		completions := make([]string, len(e.Completions))
		for i, d := range e.Completions {
			completions[i] = d.String()
		}
		resp.Habits = append(resp.Habits, habitHistoryEntryBody{
			ID: e.HabitID.String(), Name: e.Name, Archived: e.Archived, Completions: completions,
		})
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

type habitWeekEntryBody struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	CurrentStreak int      `json:"current_streak"`
	Completed     []string `json:"completed"`
}

type archivedHabitNameBody struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type weekResponse struct {
	WeekStart string                  `json:"week_start"`
	Days      []string                `json:"days"`
	Habits    []habitWeekEntryBody    `json:"habits"`
	Archived  []archivedHabitNameBody `json:"archived"`
}

// week serves GET /api/habits/week?date=<any-day-in-week> — the ISO week
// containing date, batched into one call for the "This Week" grid (docs/left.md
// Phase 6), replacing 7 individual GET /api/habits?date= calls.
func (h *Handler) week(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("date")
	if raw == "" {
		httpx.WriteError(w, r, httpx.ValidationError(map[string]string{"date": "date query parameter is required (YYYY-MM-DD)"}))
		return
	}
	date, err := timezone.ParseDate(raw)
	if err != nil {
		httpx.WriteError(w, r, httpx.ValidationError(map[string]string{"date": "must be YYYY-MM-DD"}))
		return
	}

	wv, err := h.svc.Week(r.Context(), accountID(r), date)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	days := make([]string, len(wv.Days))
	for i, d := range wv.Days {
		days[i] = d.String()
	}
	habitsOut := make([]habitWeekEntryBody, len(wv.Habits))
	for i, hb := range wv.Habits {
		completed := make([]string, len(hb.Completed))
		for j, d := range hb.Completed {
			completed[j] = d.String()
		}
		habitsOut[i] = habitWeekEntryBody{ID: hb.HabitID.String(), Name: hb.Name, CurrentStreak: hb.CurrentStreak, Completed: completed}
	}
	archivedOut := make([]archivedHabitNameBody, len(wv.Archived))
	for i, a := range wv.Archived {
		archivedOut[i] = archivedHabitNameBody{ID: a.HabitID.String(), Name: a.Name}
	}

	httpx.WriteJSON(w, http.StatusOK, weekResponse{
		WeekStart: wv.WeekStart.String(), Days: days, Habits: habitsOut, Archived: archivedOut,
	})
}

type createHabitRequest struct {
	Name       string  `json:"name"`
	CategoryID *string `json:"category_id"`
	Target     *string `json:"target"`
}

func parseCategoryID(raw *string) (*uuid.UUID, *ValidationError) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(*raw)
	if err != nil {
		return nil, &ValidationError{Fields: map[string]string{"category_id": "must be a UUID"}}
	}
	return &id, nil
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req createHabitRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	catID, verr := parseCategoryID(req.CategoryID)
	if verr != nil {
		httpx.WriteError(w, r, httpx.ValidationError(verr.Fields))
		return
	}
	habit, err := h.svc.CreateHabit(r.Context(), accountID(r), HabitInput{Name: req.Name, CategoryID: catID, Target: req.Target})
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toHabitBody(habit))
}

type updateHabitRequest struct {
	Name   string  `json:"name"`
	Target *string `json:"target"`
}

// update serves PATCH /api/habits/{id} — replaces name + target together (MX3;
// category has its own endpoint, setCategory below).
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var req updateHabitRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	habit, err := h.svc.UpdateHabit(r.Context(), accountID(r), id, req.Name, req.Target)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toHabitBody(habit))
}

type categoryRequest struct {
	CategoryID *string `json:"category_id"`
}

func (h *Handler) setCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var req categoryRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	catID, verr := parseCategoryID(req.CategoryID)
	if verr != nil {
		httpx.WriteError(w, r, httpx.ValidationError(verr.Fields))
		return
	}
	if err := h.svc.SetHabitCategory(r.Context(), accountID(r), id, catID); err != nil {
		writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
