package vtubers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	ErrBackoff = errors.New("backoff")
)

type UpdateOptions struct {
	ScraperBatchSize   int
	MaxRequestAttempts int
	GoogleAPIKey       string
	ChannelsOnly       bool
}

func (o *UpdateOptions) applyDefaults() {
	if o.ScraperBatchSize == 0 {
		o.ScraperBatchSize = 100
	}
	if o.MaxRequestAttempts == 0 {
		o.MaxRequestAttempts = 5
	}
}

type Updater struct {
	Options UpdateOptions
	Store   *Store
	Scraper *HololistScraper
}

func (u *Updater) updateHololistData(ctx context.Context) error {
	u.Scraper.Reset()

	for {
		page, err := u.Scraper.NextPosts(ctx, u.Options.ScraperBatchSize, u.Options.MaxRequestAttempts)
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

			rendered, err := u.Scraper.GetRenderedPost(ctx, meta.Link, u.Options.MaxRequestAttempts)
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

func (u *Updater) Update(ctx context.Context) error {
	u.Options.applyDefaults()

	if !u.Options.ChannelsOnly {
		err := u.updateHololistData(ctx)
		if err != nil {
			return fmt.Errorf("hololist update: %w", err)
		}
	}
	youtubeIDs, err := u.Store.GetAllScrapedYouTubeIDs(ctx)
	if err != nil {
		return fmt.Errorf("load ids: %w", err)
	}

	err = u.updateChannelData(ctx, youtubeIDs)
	if err != nil {
		return fmt.Errorf("channel data update: %w", err)
	}

	err = u.Store.LogUpdate(ctx)
	if err != nil {
		return fmt.Errorf("log update: %w", err)
	}

	return nil
}

func (u *Updater) updateChannelData(ctx context.Context, ids []string) error {
	apiKey := u.Options.GoogleAPIKey
	if apiKey == "" {
		return nil
	}

	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return err
	}

	channelsService := youtube.NewChannelsService(service)
	batches := slices.Chunk(ids, 50)

	for batch := range batches {
		channels, err := channelsService.
			List([]string{"snippet"}).
			Context(ctx).
			Id(batch...).
			Do()
		if err != nil {
			return err
		}

		for _, channel := range channels.Items {
			thumbnailURL := ""
			if channel.Snippet.Thumbnails.Default != nil {
				thumbnailURL = channel.Snippet.Thumbnails.Default.Url
			}
			c := Channel{
				ID:        channel.Id,
				Name:      channel.Snippet.Title,
				AvatarURL: thumbnailURL,
			}
			err := u.Store.CreateOrUpdateChannel(ctx, c)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
