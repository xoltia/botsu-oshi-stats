package vtubers

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrBackoff = errors.New("backoff")
)

type Updater struct {
	db      *sqlx.DB
	store   *Store
	scraper *HololistScraper
}

func OpenUpdater(ctx context.Context, db *sqlx.DB, store *Store, scraper *HololistScraper) (*Updater, error) {
	_, err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS updater (last_update TIMESTAMP)")
	if err != nil {
		return nil, err
	}
	return &Updater{
		db:      db,
		store:   store,
		scraper: scraper,
	}, nil
}

func (u *Updater) LastUpdate(ctx context.Context) (time.Time, error) {
	var result struct {
		LastUpdate time.Time `db:"last_update"`
	}
	err := u.db.GetContext(ctx, &result, "SELECT last_update FROM updater")
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return result.LastUpdate, err
}

func (u *Updater) Update(ctx context.Context) error {
	u.scraper.Reset()

	for {
		page, err := u.scraper.NextPosts(ctx, 100)
		if err != nil {
			if errors.Is(err, ErrExhaustedPosts) {
				break
			}
			return err
		}

		for _, meta := range page {
			existing, err := u.store.FindByID(ctx, meta.ID)
			if err == nil && existing.Modified == meta.Modified {
				log.Printf("Not modified: %s\n", meta.Link)
				continue
			} else if !errors.Is(err, sql.ErrNoRows) {
				return err
			}

			log.Printf("Updating: %s\n", meta.Link)
			rendered, err := u.scraper.GetRenderedPost(ctx, meta.Link)
			if err != nil {
				return err
			}
			err = u.store.CreateOrUpdate(ctx, VTuber{rendered, meta})
			if err != nil {
				return err
			}
		}
	}

	_, err := u.db.ExecContext(ctx, "INSERT INTO updater (last_update) VALUES (CURRENT_TIMESTAMP)")
	return err
}
