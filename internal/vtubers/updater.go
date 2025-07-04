package vtubers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"iter"
	"maps"
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

func (u *Updater) updateHololistData(ctx context.Context) (iter.Seq[string], error) {
	u.Scraper.Reset()

	// Using a map because it is technically possible for multiple vtubers to share a channel.
	seenIDs := map[string]struct{}{}

	for {
		page, err := u.Scraper.NextPosts(ctx, u.Options.ScraperBatchSize, u.Options.MaxRequestAttempts)
		if err != nil {
			if errors.Is(err, ErrExhaustedPosts) {
				break
			}
			return nil, err
		}

		for _, meta := range page {
			existing, err := u.Store.FindByID(ctx, meta.ID)
			if err == nil && existing.Modified == meta.Modified {
				continue
			} else if !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}

			rendered, err := u.Scraper.GetRenderedPost(ctx, meta.Link, u.Options.MaxRequestAttempts)
			if err != nil {
				return nil, err
			}
			err = u.Store.CreateOrUpdate(ctx, VTuber{rendered, meta})
			if err != nil {
				return nil, err
			}

			if rendered.YouTubeID != "" {
				seenIDs[rendered.YouTubeID] = struct{}{}
			}
		}
	}

	return maps.Keys(seenIDs), nil
}

func (u *Updater) Update(ctx context.Context) error {
	u.Options.applyDefaults()

	var (
		youtubeIDs iter.Seq[string]
		err        error
	)

	if !u.Options.ChannelsOnly {
		youtubeIDs, err = u.updateHololistData(ctx)
		if err != nil {
			return fmt.Errorf("hololist update: %w", err)
		}
	} else {
		ids, err := u.Store.GetAllScrapedYouTubeIDs(ctx)
		if err != nil {
			return fmt.Errorf("load ids: %w", err)
		}
		youtubeIDs = slices.Values(ids)
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

func (u *Updater) updateChannelData(ctx context.Context, ids iter.Seq[string]) error {
	apiKey := u.Options.GoogleAPIKey
	if apiKey == "" {
		return nil
	}

	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return err
	}

	channelsService := youtube.NewChannelsService(service)

	idSlice := slices.Collect(ids)
	batches := slices.Chunk(idSlice, 50)

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
