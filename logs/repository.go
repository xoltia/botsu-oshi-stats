package logs

import (
	"context"
	"fmt"

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

type GetRecentUserVideosParams struct {
	UserID  string
	Limit   int
	PageKey PaginationKey
}

// GetRecentUserVideos returns all unique video log entries with the most recent video information.
// The duration field is replaced with the sum of the durations and the date is the most recent date.
func (r *UserLogRepository) GetRecentUserVideos(ctx context.Context, params GetRecentUserVideosParams) ([]VideoInfo, error) {
	const queryWithKey = `
		SELECT DISTINCT ON (meta->>'video_id') meta
		FROM activities
		WHERE
			user_id = $1 AND
			date < $3 AND
			id > $4 AND
			media_type = 'video' AND
			meta->>'platform' = 'youtube' AND
			meta->>'video_id' IS NOT NULL
		ORDER BY date DESC, id
		LIMIT $2
	`

	const queryNoKey = `
		SELECT DISTINCT ON (meta->>'video_id') meta
		FROM activities
		WHERE
			user_id = $1 AND
			media_type = 'video' AND
			meta->>'platform' = 'youtube' AND
			meta->>'video_id' IS NOT NULL
		ORDER BY date DESC, id
		LIMIT $2
	`

	args := []any{
		params.UserID,
		params.Limit,
		params.PageKey.Date,
		params.PageKey.ID,
	}

	query := queryWithKey
	if params.PageKey.IsZero() {
		query = queryNoKey
	}

	rows, err := r.db.QueryxContext(
		ctx,
		query,
		args...,
	)

	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	logs := make([]VideoInfo, 0, params.Limit)
	for rows.Next() {
		var log Log
		err = rows.StructScan(&log)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		logs = append(logs, log.Video)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("next: %w", err)
	}

	return logs, nil
}
