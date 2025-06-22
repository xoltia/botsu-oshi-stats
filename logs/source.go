package logs

import (
	"context"
	"math"
	"time"

	"github.com/jmoiron/sqlx"
)

type Source struct {
	db *sqlx.DB
}

func NewSource(db *sqlx.DB) *Source {
	return &Source{db: db.Unsafe()}
}

// PaginationKey contains the info needed to get
// the next page of values when querying logs.
// The zero value starts from the beginning.
type PaginationKey struct {
	ID   int64
	Date time.Time
}

// PaginationKeyFromSet returns a key based on a sorted set.
// Returns `nil` for an empty set, so be sure to check whether
// the last result was empty before calling again.
func PaginationKeyFromSet(set []Log) *PaginationKey {
	if len(set) == 0 {
		return nil
	}

	return &PaginationKey{
		ID:   set[len(set)-1].ID,
		Date: set[len(set)-1].Date,
	}
}

// GetByUser returns up to `limit` (0 for no limit) logs for a given `userID` in descending order by date.
// Uses `key` to determine the offset. Use `nil` to start from the beginning.
func (s *Source) GetByUser(ctx context.Context, userID string, key *PaginationKey, limit int) ([]Log, error) {
	if limit < 1 {
		limit = math.MaxInt
	}

	lastID := int64(math.MaxInt64)
	lastDate := time.Now()
	if key != nil {
		lastID = key.ID
		lastDate = key.Date
	}

	logs := make([]Log, 0, limit)
	rows, err := s.db.Queryx(`
		SELECT * FROM activities
		WHERE media_type = 'video'
			AND meta->>'platform' = 'youtube'
			AND user_id = $1
			AND id < $2
			AND date < $3
			AND deleted_at IS NULL
		ORDER BY date DESC, id DESC
		LIMIT $4
		`, userID, lastID, lastDate, limit)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var log Log
		err := rows.StructScan(&log)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, err

}
