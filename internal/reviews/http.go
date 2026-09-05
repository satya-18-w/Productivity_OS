package reviews

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/satya-18-w/productivity-os/internal/platform/httpx"
	"github.com/satya-18-w/productivity-os/internal/platform/reqctx"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

// Handler serves the review HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler builds the reviews handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Protector wraps a handler with auth (write also adds CSRF). cmd/server supplies
// the account module's middleware.
type Protector func(http.HandlerFunc) http.Handler

// Mount registers the review routes. write must enforce auth + CSRF; read only
// auth.
func (h *Handler) Mount(mux *http.ServeMux, write, read Protector) {
	mux.Handle("GET /api/reviews/daily", read(h.getDaily))
	mux.Handle("PUT /api/reviews/daily", write(h.putDaily))
	mux.Handle("GET /api/reviews/weekly", read(h.getWeekly))
	mux.Handle("PUT /api/reviews/weekly", write(h.putWeekly))
}

func accountID(r *http.Request) uuid.UUID {
	id, _ := reqctx.IdentityFrom(r.Context())
	return id.AccountID
}

type promptBody struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}

func promptBodies(prompts []Prompt) []promptBody {
	out := make([]promptBody, len(prompts))
	for i, p := range prompts {
		out[i] = promptBody(p)
	}
	return out
}

type answersRequest struct {
	Answers Answers `json:"answers"`
}

func queryDate(w http.ResponseWriter, r *http.Request) (timezone.Date, bool) {
	raw := r.URL.Query().Get("date")
	if raw == "" {
		httpx.WriteError(w, r, httpx.ValidationError(map[string]string{
			"date": "date query parameter is required (YYYY-MM-DD)",
		}))
		return timezone.Date{}, false
	}
	d, err := timezone.ParseDate(raw)
	if err != nil {
		httpx.WriteError(w, r, httpx.ValidationError(map[string]string{"date": "must be YYYY-MM-DD"}))
		return timezone.Date{}, false
	}
	return d, true
}

func queryYearWeek(w http.ResponseWriter, r *http.Request) (year, week int, ok bool) {
	q := r.URL.Query()
	fields := map[string]string{}

	year, yErr := strconv.Atoi(q.Get("year"))
	if yErr != nil {
		fields["year"] = "must be an integer (ISO year)"
	}
	week, wErr := strconv.Atoi(q.Get("week"))
	if wErr != nil {
		fields["week"] = "must be an integer (ISO week, 1-53)"
	}
	if len(fields) > 0 {
		httpx.WriteError(w, r, httpx.ValidationError(fields))
		return 0, 0, false
	}
	return year, week, true
}

type dailyReviewBody struct {
	Date      string       `json:"date"`
	Prompts   []promptBody `json:"prompts"`
	Answers   Answers      `json:"answers"`
	UpdatedAt string       `json:"updated_at,omitempty"`
}

func (h *Handler) getDaily(w http.ResponseWriter, r *http.Request) {
	date, ok := queryDate(w, r)
	if !ok {
		return
	}
	rev, err := h.svc.GetDaily(r.Context(), accountID(r), date)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, dailyReviewBody{
		Date: date.String(), Prompts: promptBodies(DailyPrompts), Answers: rev.Answers, UpdatedAt: rev.UpdatedAt,
	})
}

func (h *Handler) putDaily(w http.ResponseWriter, r *http.Request) {
	date, ok := queryDate(w, r)
	if !ok {
		return
	}
	var req answersRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	if _, err := h.svc.SaveDaily(r.Context(), accountID(r), date, req.Answers); err != nil {
		writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type weeklyReviewBody struct {
	ISOYear   int          `json:"iso_year"`
	ISOWeek   int          `json:"iso_week"`
	Prompts   []promptBody `json:"prompts"`
	Answers   Answers      `json:"answers"`
	UpdatedAt string       `json:"updated_at,omitempty"`
}

func (h *Handler) getWeekly(w http.ResponseWriter, r *http.Request) {
	year, week, ok := queryYearWeek(w, r)
	if !ok {
		return
	}
	rev, err := h.svc.GetWeekly(r.Context(), accountID(r), year, week)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, weeklyReviewBody{
		ISOYear: year, ISOWeek: week, Prompts: promptBodies(WeeklyPrompts), Answers: rev.Answers, UpdatedAt: rev.UpdatedAt,
	})
}

func (h *Handler) putWeekly(w http.ResponseWriter, r *http.Request) {
	year, week, ok := queryYearWeek(w, r)
	if !ok {
		return
	}
	var req answersRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	if _, err := h.svc.SaveWeekly(r.Context(), accountID(r), year, week, req.Answers); err != nil {
		writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var verr *ValidationError
	if errors.As(err, &verr) {
		httpx.WriteError(w, r, httpx.ValidationError(verr.Fields))
		return
	}
	httpx.WriteError(w, r, err)
}
