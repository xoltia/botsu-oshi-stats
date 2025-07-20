package vtubers

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"strings"

	"github.com/xoltia/botsu-oshi-stats/logs"
	"github.com/xoltia/botsu-oshi-stats/multimatch"
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
		if acceptableOriginalName(entry.OriginalName) {
			cleanName := strings.Replace(entry.OriginalName, "・", "", -1)
			builder.AddString(cleanName, entry.ID)
		}
		if acceptableEnglishName(entry.EnglishName) {
			builder.AddString(entry.EnglishName, entry.ID)
		}
	}

	matcher := builder.Build()
	return &Detector{
		dictionary: matcher,
		store:      s,
	}, nil
}

// Filter for English names that are too likely to have false positives.
func acceptableEnglishName(s string) bool {
	return strings.IndexByte(s, ' ') != -1
}

// Filter for non-English (hopefully Japanese) names that may have
// false positives (i.e. 叶)
func acceptableOriginalName(s string) bool {
	runes := []rune(s)
	hasNonKana := strings.ContainsFunc(s, func(r rune) bool {
		katakana := r >= 0x30A0 && r <= 0x30FF
		hiragana := r >= 0x3041 && r <= 0x3096
		return !(katakana || hiragana)
	})
	return (hasNonKana && len(runes) > 1) || len(runes) >= 5
}

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
	// TODO: Fix this, vtubers can share channels
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

	// TODO: Also check hashtags

	result.LinkedChannel = result.All[linkedStart:linkedEnd]
	result.NameText = result.All[linkedEnd:]
	return result, nil
}
