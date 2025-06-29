package vtubers

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrBackoff = errors.New("backoff")
)

type UpdateOptions struct {
	BatchSize          int
	MaxRequestAttempts int
}

func (o *UpdateOptions) applyDefaults() {
	if o.BatchSize == 0 {
		o.BatchSize = 100
	}
	if o.MaxRequestAttempts == 0 {
		o.MaxRequestAttempts = 5
	}
}

type UpdateOption func(o *UpdateOptions)

func WithBatchSize(size int) UpdateOption {
	return func(o *UpdateOptions) {
		o.BatchSize = size
	}
}

// Number of times each request can be retried before
// returning `ErrBackoff`.
func WithMaxRequestAttempts(attempts int) UpdateOption {
	return func(o *UpdateOptions) {
		o.MaxRequestAttempts = attempts
	}
}

type Updater struct {
	Store   *Store
	Scraper *HololistScraper
}

func (u *Updater) Update(ctx context.Context, options ...UpdateOption) error {
	u.Scraper.Reset()

	opts := UpdateOptions{}
	for _, f := range options {
		f(&opts)
	}
	opts.applyDefaults()

	for {
		page, err := u.Scraper.NextPosts(ctx, opts.BatchSize, opts.MaxRequestAttempts)
		if err != nil {
			if errors.Is(err, ErrExhaustedPosts) {
				break
			}
			return err
		}

		for _, meta := range page {
			existing, err := u.Store.FindByID(ctx, meta.ID)
			if err == nil && existing.Modified == meta.Modified {
				continue
			} else if !errors.Is(err, sql.ErrNoRows) {
				return err
			}

			rendered, err := u.Scraper.GetRenderedPost(ctx, meta.Link, opts.MaxRequestAttempts)
			if err != nil {
				return err
			}
			err = u.Store.CreateOrUpdate(ctx, VTuber{rendered, meta})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
