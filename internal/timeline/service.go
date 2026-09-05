package timeline

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/satya-18-w/productivity-os/internal/timeline/timelinedb"
)

type service struct {
	pool  *pgxpool.Pool
	q     *timelinedb.Queries
	zone  AccountZone
	cats  CategoryStore
	tasks TaskChecker
}

// NewService builds the timeline service over a connection pool. zone resolves an
// account's timezone; cats is the categories module, used to validate and label
// the category a block is assigned to (wired by cmd/server — ADR-0009); tasks
// validates a task a block links to and resolves inherited categories (MX-TL).
func NewService(pool *pgxpool.Pool, zone AccountZone, cats CategoryStore, tasks TaskChecker) Service {
	return &service{pool: pool, q: timelinedb.New(pool), zone: zone, cats: cats, tasks: tasks}
}

func toPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func fromPgUUID(v pgtype.UUID) *uuid.UUID {
	if !v.Valid {
		return nil
	}
	id := uuid.UUID(v.Bytes)
	return &id
}
