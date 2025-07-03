package logs

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type UserLogRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *UserLogRepository {
	return &UserLogRepository{db}
}

type LogSet struct {
	rows *sqlx.Rows
}

func (s LogSet) Next() bool {
	return s.rows.Next()
}

func (s LogSet) Err() error {
	return s.rows.Err()
}

func (s LogSet) Scan() (Log, error) {
	var log Log
	err := s.rows.StructScan(&log)
	return log, err
}

func (s LogSet) Close() error {
	return s.rows.Close()
}

func (r *UserLogRepository) querySet(ctx context.Context, query string, args ...any) (logs LogSet, err error) {
	rows, err := r.db.QueryxContext(ctx, query, args...)
	logs = LogSet{rows: rows}
	return
}

func (r *UserLogRepository) GetAll(ctx context.Context) (logs LogSet, err error) {
	return r.querySet(ctx, `
		SELECT id, user_id, date, duration, meta
		FROM activities
		WHERE media_type = 'video' AND meta->>'platform' = 'youtube';
	`)
}
