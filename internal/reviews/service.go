package reviews

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
	"github.com/satya-18-w/productivity-os/internal/reviews/reviewsdb"
)

type service struct {
	q *reviewsdb.Queries
}

// NewService builds the reviews service over a connection pool.
func NewService(pool *pgxpool.Pool) Service {
	return &service{q: reviewsdb.New(pool)}
}

func (s *service) GetDaily(ctx context.Context, accountID uuid.UUID, date timezone.Date) (DailyReview, error) {
	row, err := s.q.GetDailyReview(ctx, reviewsdb.GetDailyReviewParams{
		AccountID: accountID, OnDate: pgDate(date),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DailyReview{Date: date, Answers: Answers{}}, nil
		}
		return DailyReview{}, fmt.Errorf("get daily review: %w", err)
	}
	answers, err := decodeAnswers(row.Answers)
	if err != nil {
		return DailyReview{}, err
	}
	return DailyReview{Date: date, Answers: answers, UpdatedAt: rfc3339(row.UpdatedAt)}, nil
}

func (s *service) SaveDaily(ctx context.Context, accountID uuid.UUID, date timezone.Date, answers Answers) (DailyReview, error) {
	clean := filterKnown(DailyPrompts, answers)
	encoded, err := json.Marshal(clean)
	if err != nil {
		return DailyReview{}, fmt.Errorf("encode answers: %w", err)
	}
	updatedAt, err := s.q.UpsertDailyReview(ctx, reviewsdb.UpsertDailyReviewParams{
		AccountID: accountID, OnDate: pgDate(date), Answers: encoded,
	})
	if err != nil {
		return DailyReview{}, fmt.Errorf("save daily review: %w", err)
	}
	return DailyReview{Date: date, Answers: clean, UpdatedAt: rfc3339(updatedAt)}, nil
}

func (s *service) GetWeekly(ctx context.Context, accountID uuid.UUID, isoYear, isoWeek int) (WeeklyReview, error) {
	if verr := validateYearWeek(isoYear, isoWeek); verr != nil {
		return WeeklyReview{}, verr
	}
	row, err := s.q.GetWeeklyReview(ctx, reviewsdb.GetWeeklyReviewParams{
		AccountID: accountID,
		IsoYear:   int32(isoYear), //nolint:gosec // G115: validateYearWeek bounds isoYear to [1,9999] above
		IsoWeek:   int32(isoWeek), //nolint:gosec // G115: validateYearWeek bounds isoWeek to [1,53] above
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return WeeklyReview{ISOYear: isoYear, ISOWeek: isoWeek, Answers: Answers{}}, nil
		}
		return WeeklyReview{}, fmt.Errorf("get weekly review: %w", err)
	}
	answers, err := decodeAnswers(row.Answers)
	if err != nil {
		return WeeklyReview{}, err
	}
	return WeeklyReview{ISOYear: isoYear, ISOWeek: isoWeek, Answers: answers, UpdatedAt: rfc3339(row.UpdatedAt)}, nil
}

func (s *service) SaveWeekly(ctx context.Context, accountID uuid.UUID, isoYear, isoWeek int, answers Answers) (WeeklyReview, error) {
	if verr := validateYearWeek(isoYear, isoWeek); verr != nil {
		return WeeklyReview{}, verr
	}
	clean := filterKnown(WeeklyPrompts, answers)
	encoded, err := json.Marshal(clean)
	if err != nil {
		return WeeklyReview{}, fmt.Errorf("encode answers: %w", err)
	}
	updatedAt, err := s.q.UpsertWeeklyReview(ctx, reviewsdb.UpsertWeeklyReviewParams{
		AccountID: accountID,
		IsoYear:   int32(isoYear), //nolint:gosec // G115: validateYearWeek bounds isoYear to [1,9999] above
		IsoWeek:   int32(isoWeek), //nolint:gosec // G115: validateYearWeek bounds isoWeek to [1,53] above
		Answers:   encoded,
	})
	if err != nil {
		return WeeklyReview{}, fmt.Errorf("save weekly review: %w", err)
	}
	return WeeklyReview{ISOYear: isoYear, ISOWeek: isoWeek, Answers: clean, UpdatedAt: rfc3339(updatedAt)}, nil
}

func (s *service) ListDaily(ctx context.Context, accountID uuid.UUID) ([]DailyReview, error) {
	rows, err := s.q.ListAllDailyReviews(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("list all daily reviews: %w", err)
	}
	out := make([]DailyReview, len(rows))
	for i, r := range rows {
		answers, err := decodeAnswers(r.Answers)
		if err != nil {
			return nil, err
		}
		out[i] = DailyReview{
			Date:      timezone.Date{Year: r.OnDate.Time.Year(), Month: r.OnDate.Time.Month(), Day: r.OnDate.Time.Day()},
			Answers:   answers,
			UpdatedAt: rfc3339(r.UpdatedAt),
		}
	}
	return out, nil
}

func (s *service) ListWeekly(ctx context.Context, accountID uuid.UUID) ([]WeeklyReview, error) {
	rows, err := s.q.ListAllWeeklyReviews(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("list all weekly reviews: %w", err)
	}
	out := make([]WeeklyReview, len(rows))
	for i, r := range rows {
		answers, err := decodeAnswers(r.Answers)
		if err != nil {
			return nil, err
		}
		out[i] = WeeklyReview{
			ISOYear:   int(r.IsoYear),
			ISOWeek:   int(r.IsoWeek),
			Answers:   answers,
			UpdatedAt: rfc3339(r.UpdatedAt),
		}
	}
	return out, nil
}

const (
	minISOYear = 1
	maxISOYear = 9999
)

func validateYearWeek(isoYear, isoWeek int) *ValidationError {
	fields := map[string]string{}
	if isoYear < minISOYear || isoYear > maxISOYear {
		fields["iso_year"] = "must be between 1 and 9999"
	}
	if isoWeek < 1 || isoWeek > 53 {
		fields["iso_week"] = "must be between 1 and 53"
	}
	if len(fields) > 0 {
		return &ValidationError{Fields: fields}
	}
	return nil
}

// filterKnown drops any answer key not in the fixed prompt set — the prompt set
// is not user-editable, so an unknown key can only be a stale client or a bad
// request, never a legitimate new prompt.
func filterKnown(known []Prompt, in Answers) Answers {
	keys := make(map[string]struct{}, len(known))
	for _, p := range known {
		keys[p.Key] = struct{}{}
	}
	out := Answers{}
	for k, v := range in {
		if _, ok := keys[k]; ok {
			out[k] = v
		}
	}
	return out
}

func decodeAnswers(raw []byte) (Answers, error) {
	if len(raw) == 0 {
		return Answers{}, nil
	}
	var out Answers
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode answers: %w", err)
	}
	if out == nil {
		out = Answers{}
	}
	return out, nil
}

func pgDate(d timezone.Date) pgtype.Date {
	return pgtype.Date{Time: time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC), Valid: true}
}

func rfc3339(ts pgtype.Timestamptz) string {
	if !ts.Valid {
		return ""
	}
	return ts.Time.UTC().Format(time.RFC3339)
}
