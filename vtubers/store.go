package vtubers

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Store struct {
	db *sqlx.DB
}

func CreateStore(ctx context.Context, db *sqlx.DB) (*Store, error) {
	_, err := db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS vtubers (
				youtube_id     TEXT NOT NULL,
				youtube_handle TEXT NOT NULL,
				original_name  TEXT NOT NULL,
				picture_url    TEXT NOT NULL,
				english_name   TEXT NOT NULL,
				oshi_mark      TEXT NOT NULL,
				zodiac         TEXT NOT NULL,
				affiliation    TEXT NOT NULL,
				birthday       TEXT NOT NULL,
				debut_date     TEXT NOT NULL,
				gender         TEXT NOT NULL,
				fanbase        TEXT NOT NULL,
				status         TEXT NOT NULL,
				id             INTEGER PRIMARY KEY,
				link           TEXT NOT NULL NOT NULL,
				modified       TEXT NOT NULL
			);

			CREATE TABLE IF NOT EXISTS vtuber_channels (
				id         TEXT NOT NULL PRIMARY KEY,
				name       TEXT NOT NULL,
				avatar_url TEXT NOT NULL
			);

			CREATE TABLE IF NOT EXISTS update_history (
				timestamp TIMESTAMP NOT NULL
			);
		`)
	if err != nil {
		return nil, err
	}
	store := &Store{db}
	return store, nil
}

func (s *Store) CreateOrUpdateChannel(ctx context.Context, c Channel) error {
	_, err := s.db.NamedExecContext(ctx, `
			INSERT INTO vtuber_channels (
				id,
				name,
				avatar_url
			)
			VALUES (
				:id,
				:name,
				:avatar_url
			)
			ON CONFLICT (id) DO UPDATE
			SET 
				id = :id,
				name = :name,
				avatar_url = :avatar_url
		`, c)
	return err
}

func (s *Store) CreateOrUpdate(ctx context.Context, v VTuber) error {
	_, err := s.db.NamedExecContext(ctx, `
			INSERT INTO vtubers (
				youtube_id,
				youtube_handle,
				original_name,
				english_name,
				oshi_mark,
				zodiac,
				affiliation,
				birthday,
				debut_date,
				gender,
				fanbase,
				status,
				id,
				link,
				modified,
				picture_url
			)
			VALUES (
				:youtube_id,
				:youtube_handle,
				:original_name,
				:english_name,
				:oshi_mark,
				:zodiac,
				:affiliation,
				:birthday,
				:debut_date,
				:gender,
				:fanbase,
				:status,
				:id,
				:link,
				:modified,
				:picture_url
			)
			ON CONFLICT (id) DO UPDATE
			SET 
				youtube_id = :youtube_id,
				youtube_handle = :youtube_handle,
				original_name = :original_name,
				english_name = :english_name,
				oshi_mark = :oshi_mark,
				zodiac = :zodiac,
				affiliation = :affiliation,
				birthday = :birthday,
				debut_date = :debut_date,
				gender = :gender,
				fanbase = :fanbase,
				status = :status,
				id = :id,
				link = :link,
				modified = :modified,
				picture_url = :picture_url
		`, v)
	return err
}

func (s *Store) FindByID(ctx context.Context, id int) (v VTuber, err error) {
	err = s.db.GetContext(ctx, &v, "SELECT * FROM vtubers WHERE id = $1", id)
	return
}

func (s *Store) FindByName(ctx context.Context, name string) (v VTuber, err error) {
	err = s.db.GetContext(ctx, &v, "SELECT * FROM vtubers WHERE original_name = $1 or english_name = $1", name)
	return
}

func (s *Store) FindByYouTubeID(ctx context.Context, id string) (v VTuber, err error) {
	err = s.db.GetContext(ctx, &v, "SELECT * FROM vtubers WHERE youtube_id = $1", id)
	return
}

func (s *Store) FindByYouTubeHandle(ctx context.Context, handle string) (v VTuber, err error) {
	err = s.db.GetContext(ctx, &v, "SELECT * FROM vtubers WHERE youtube_handle = $1 COLLATE NOCASE", handle)
	return
}

func (s *Store) FindChannelByID(ctx context.Context, id string) (c Channel, err error) {
	err = s.db.GetContext(ctx, &c, "SELECT * FROM vtuber_channels WHERE id = $1", id)
	return
}

func (s *Store) GetAllScrapedYouTubeIDs(ctx context.Context) (ids []string, err error) {
	rows, err := s.db.QueryxContext(ctx, "SELECT DISTINCT youtube_id FROM vtubers WHERE youtube_id != ''")
	if err != nil {
		err = fmt.Errorf("query: %w", err)
		return
	}

	for rows.Next() {
		var row struct {
			ID string `db:"youtube_id"`
		}
		if err = rows.StructScan(&row); err != nil {
			err = fmt.Errorf("scan: %w", err)
			return
		}
		ids = append(ids, row.ID)
	}

	if err = rows.Err(); err != nil {
		err = fmt.Errorf("iteration: %w", err)
	}
	return
}

type Names struct {
	ID           int    `db:"id"`
	OriginalName string `db:"original_name"`
	EnglishName  string `db:"english_name"`
}

func (s *Store) GetAllNames(ctx context.Context) (names []Names, err error) {
	rows, err := s.db.QueryxContext(ctx, "SELECT id, original_name, english_name FROM vtubers")
	if err != nil {
		return nil, err
	}

	names = make([]Names, 0)
	for rows.Next() {
		var n Names
		if err = rows.StructScan(&n); err != nil {
			err = fmt.Errorf("scan: %w", err)
			return
		}
		names = append(names, n)
	}

	if err = rows.Err(); err != nil {
		err = fmt.Errorf("iteration: %w", err)
	}
	return
}

func (s *Store) LogUpdate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO update_history (timestamp) VALUES (CURRENT_TIMESTAMP)")
	return err
}

func (s *Store) LastUpdate(ctx context.Context) (time.Time, error) {
	var row struct {
		Timestamp time.Time `db:"timestamp"`
	}
	err := s.db.GetContext(ctx, &row, "SELECT MAX(timestamp) FROM updates")
	return row.Timestamp, err
}
