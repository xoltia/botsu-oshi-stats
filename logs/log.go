package logs

import (
	"encoding/json"
	"errors"
	"time"
)

type Log struct {
	ID        int64         `json:"id" db:"id"`
	Duration  time.Duration `json:"duration" db:"duration"`
	UserID    string        `json:"user_id" db:"user_id"`
	MediaType string        `json:"media_type" db:"media_type"`
	Date      time.Time     `json:"date" db:"date"`
	DeletedAt *time.Time    `json:"deleted_at" db:"deleted_at"`
	VideoMeta VideoMeta     `json:"meta" db:"meta"`
}

type VideoMeta struct {
	Platform       string   `json:"platform" db:"platform"`
	ChannelID      string   `json:"channel_id" db:"channel_id"`
	ChannelHandle  string   `json:"channel_handle" db:"channel_handle"`
	ChannelName    string   `json:"channel_name" db:"channel_name"`
	VideoID        string   `json:"video_id" db:"video_id"`
	Title          string   `json:"video_title" db:"video_title"`
	LinkedChannels []string `json:"linked_channels" db:"linked_channels"`
	LinkedVideos   []string `json:"linked_videos" db:"linked_videos"`
	HashTags       []string `json:"hashtags" db:"hashtags"`
	Thumbnail      string   `json:"thumbnail" db:"thumbnail"`
}

func (m *VideoMeta) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("failed type assertion to []byte")
	}
	return json.Unmarshal(b, &m)
}
