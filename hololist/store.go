package hololist

import (
	"context"

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
			)
		`)
	if err != nil {
		return nil, err
	}
	store := &Store{db}
	return store, nil
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
	err = s.db.GetContext(ctx, &v, "SELECT * FROM vtubers WHERE youtube_handle = $1", handle)
	return
}
