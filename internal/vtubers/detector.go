package vtubers

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"strings"

	"github.com/xoltia/botsu-oshi-stats/internal/logs"
	"github.com/xoltia/botsu-oshi-stats/internal/multimatch"
)

type Detector struct {
	dictionary multimatch.Matcher
	store      *Store
}

func CreateDetector(ctx context.Context, s *Store) (*Detector, error) {
	names, err := s.GetAllNames(ctx)
	if err != nil {
		return nil, err
	}

	builder := multimatch.Builder{}
	for _, entry := range names {
		if entry.OriginalName != "" {
			builder.AddString(entry.OriginalName, entry.ID)
		}
		if entry.EnglishName != "" {
			builder.AddString(entry.EnglishName, entry.ID)
		}
	}

	matcher := builder.Build()
	return &Detector{
		dictionary: matcher,
		store:      s,
	}, nil
}

type DetectionType int

// DetectionResult contains the result of searching for hints of a vtubers
// presence from video metadata. A single vtuber is only detected once per
// type of detection, with more significant methods being attempted first.
// The complete order being: Primary Channel > Linked Channel > Name Search.
type DetectionResult struct {
	// All VTubers detected.
	All []VTuber
	// A single vtuber instance that matches the video uploader's channel ID.
	PrimaryChannel *VTuber
	// All channels linked in the YouTube description using handles or links.
	LinkedChannel []VTuber
	// Names attributes through well formated titles.
	// TitleAttribution []VTuber
	// Names found anywhere else in the video text.
	NameText []VTuber
}

func appendUnique(slice []VTuber, element VTuber) []VTuber {
	contains := func(vtuber VTuber) bool {
		return vtuber.ID == element.ID
	}
	if slices.ContainsFunc(slice, contains) {
		return slice
	} else {
		return append(slice, element)
	}
}

func (d *Detector) Detect(ctx context.Context, log logs.Log) (DetectionResult, error) {
	result := DetectionResult{}
	vtuber, err := d.store.FindByYouTubeID(ctx, log.Video.ChannelID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return DetectionResult{}, err
	} else if err == nil {
		result.All = append(result.All, vtuber)
		result.PrimaryChannel = &result.All[0]
	}

	linkedStart := len(result.All)
	for _, link := range log.Video.LinkedChannels {
		if strings.HasPrefix(link, "UC") {
			vtuber, err = d.store.FindByYouTubeID(ctx, link)
		} else if strings.HasPrefix(link, "@") {
			vtuber, err = d.store.FindByYouTubeHandle(ctx, link)
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return DetectionResult{}, err
		} else if err == nil {
			result.All = appendUnique(result.All, vtuber)
		}
	}
	// First index after end of linked channels subslice.
	linkedEnd := len(result.All)

	ids := d.dictionary.SearchString(log.Video.Title)
	for id := range ids {
		vtuber, err := d.store.FindByID(ctx, id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return DetectionResult{}, err
		} else if err == nil {
			result.All = appendUnique(result.All, vtuber)
		}
	}

	result.LinkedChannel = result.All[linkedStart:linkedEnd]
	result.NameText = result.All[linkedEnd:]
	return result, nil
}
