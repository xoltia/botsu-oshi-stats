package index

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/xoltia/botsu-oshi-stats/vtubers"
)

type IndexedVideoHistory struct {
	VideoID  string        `db:"video_id"`
	UserID   string        `db:"user_id"`
	Date     time.Time     `db:"date"`
	Duration time.Duration `db:"duration"`
}

type IndexedVideoVTuber struct {
	LogID    int    `db:"log_id"`
	VideoID  string `db:"video_id"`
	UserID   string `db:"user_id"`
	VTuberID int    `db:"vtuber_id"`
}

type IndexedVideoRepository struct {
	db *sqlx.DB
}

func CreateIndexedVideoRepository(ctx context.Context, db *sqlx.DB) (*IndexedVideoRepository, error) {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS video_vtubers (
			video_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			vtuber_id TEXT NOT NULL,

			FOREIGN KEY (vtuber_id) REFERENCES vtubers(id),
			PRIMARY KEY (video_id, user_id, vtuber_id)	
		);

		CREATE TABLE IF NOT EXISTS video_history (
			video_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			log_id INTEGER NOT NULL PRIMARY KEY,
			date TEXT NOT NULL,
			duration INTEGER NOT NULL
		);
	`)
	if err != nil {
		return nil, err
	}

	return &IndexedVideoRepository{db}, nil
}

func (r *IndexedVideoRepository) GetVTubersForVideo(
	ctx context.Context,
	userID string,
	videoID string,
) ([]vtubers.VTuber, error) {
	result := make([]vtubers.VTuber, 0)
	rows, err := r.db.QueryxContext(ctx, `
		SELECT vtubers.* FROM vtubers
		LEFT JOIN video_vtubers
		ON video_vtubers.vtuber_id = vtubers.id
		WHERE video_vtubers.user_id = ? and video_vtubers.video_id = ?
	`, userID, videoID)

	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	for rows.Next() {
		var row vtubers.VTuber
		err := rows.StructScan(&row)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		result = append(result, row)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("next: %w", err)
	}

	return result, nil
}

func (r *IndexedVideoRepository) InsertVideoVTuber(
	ctx context.Context,
	userID string,
	videoID string,
	vtuberID int,
) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO video_vtubers (user_id, video_id, vtuber_id)
		VALUES (?, ?, ?)
		ON CONFLICT DO NOTHING
	`, userID, videoID, vtuberID)
	return err
}

func (r *IndexedVideoRepository) InsertVideoHistory(
	ctx context.Context,
	userID string,
	videoID string,
	logID int,
	date time.Time,
	duration time.Duration,
) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO video_history (user_id, video_id, log_id, date, duration)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT DO NOTHING
	`, userID, videoID, logID, date.UTC(), duration)
	return err
}

func (r *IndexedVideoRepository) GetTopVTubersByAppearenceCount(
	ctx context.Context,
	userID string,
	start, end time.Time,
	limit int,
) ([]vtubers.VTuber, error) {
	rows, err := r.db.Unsafe().QueryxContext(ctx, `
		SELECT vtb.*, COUNT(*) AS appearances
		FROM video_history vh
		JOIN video_vtubers vv
		ON vh.video_id = vv.video_id AND vh.user_id = vv.user_id
		JOIN vtubers vtb
		ON vv.vtuber_id = vtb.id
		WHERE vh.user_id = ?
		      AND date(vh.date) BETWEEN date(?) AND date(?)
		GROUP BY vtb.id
		ORDER BY appearances DESC
		LIMIT ?
	`, userID, start.UTC(), end.UTC(), limit)

	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	result := make([]vtubers.VTuber, 0)
	for rows.Next() {
		var row vtubers.VTuber
		err := rows.StructScan(&row)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		result = append(result, row)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("next: %w", err)
	}

	return result, nil
}
