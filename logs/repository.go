package logs

import (
	"context"
	"fmt"
	"time"

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
		WHERE media_type = 'video' AND meta->>'platform' = 'youtube' AND deleted_at IS NULL;
	`)
}

type GetRecentUserVideosParams struct {
	UserID  string
	Limit   int
	PageKey PaginationKey
}

// GetRecentUserVideos returns all unique video log entries with the most recent video information.
// The pagination key returned may be used to get the next batch of videos. Must check for the zero
// key which indicates no more videos. Passing the zero key back in will start from the beginning.
func (r *UserLogRepository) GetRecentUserVideos(
	ctx context.Context,
	params GetRecentUserVideosParams,
) ([]VideoInfo, PaginationKey, error) {
	const queryWithKey = `
		SELECT date, id, meta
		FROM (
			SELECT DISTINCT ON (meta->>'video_id') date, id, meta
			FROM activities
			WHERE
				user_id = $1 AND
				((date < $3) OR (date = $3 AND id < $4)) AND
				media_type = 'video' AND
				meta->>'platform' = 'youtube' AND
				meta->>'video_id' IS NOT NULL AND
				deleted_at IS NULL
			ORDER BY meta->>'video_id', date DESC
		)
		ORDER BY date DESC, id DESC
		LIMIT $2
	`

	const queryNoKey = `
		SELECT date, id, meta
		FROM (
			SELECT DISTINCT ON (meta->>'video_id') date, id, meta
			FROM activities
			WHERE
				user_id = $1 AND
				media_type = 'video' AND
				meta->>'platform' = 'youtube' AND
				meta->>'video_id' IS NOT NULL AND
				deleted_at is NULL
			ORDER BY meta->>'video_id', date DESC
		)
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
		args = args[:2]
	}

	rows, err := r.db.QueryxContext(
		ctx,
		query,
		args...,
	)

	if err != nil {
		return nil, PaginationKey{}, fmt.Errorf("query: %w", err)
	}

	var log Log
	logs := make([]VideoInfo, 0, params.Limit)
	for rows.Next() {
		err = rows.StructScan(&log)
		if err != nil {
			return nil, PaginationKey{}, fmt.Errorf("scan: %w", err)
		}
		logs = append(logs, log.Video)
	}

	if err := rows.Err(); err != nil {
		return nil, PaginationKey{}, fmt.Errorf("next: %w", err)
	}

	key := PaginationKey{}
	if len(logs) > 0 {
		key.Date = log.Date
		key.ID = uint64(log.ID)
	}

	return logs, key, nil
}

func (r *UserLogRepository) GetTotalVideoWatchTime(ctx context.Context, userID string, videoID string) (time.Duration, error) {
	var row struct {
		TotalDuration time.Duration `db:"total_duration"`
	}

	err := r.db.GetContext(ctx, &row, `
		SELECT COALESCE(SUM(duration), 0) AS total_duration
		FROM activities
		WHERE media_type = 'video' AND
			  user_id = $1 AND
			  meta->>'platform' = 'youtube' AND
			  meta->>'video_id' = $2 AND
			  deleted_at is NULL
	`, userID, videoID)

	return row.TotalDuration, err
}
