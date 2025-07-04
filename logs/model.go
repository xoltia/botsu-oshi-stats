package logs

import (
	"encoding/json"
	"fmt"
	"time"
)

type VideoInfo struct {
	ID             string        `json:"video_id"`
	Title          string        `json:"video_title"`
	ChannelID      string        `json:"channel_id"`
	ChannelHandle  string        `json:"channel_handle"`
	LinkedChannels []string      `json:"linked_channels"`
	Duration       time.Duration `json:"video_duration"`
}

func (v *VideoInfo) Scan(value any) error {
	data, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected byte array, got %T", value)
	}
	return json.Unmarshal(data, v)
}

type Log struct {
	ID       int           `db:"id"`
	UserID   string        `db:"user_id"`
	Date     time.Time     `db:"date"`
	Duration time.Duration `db:"duration"`
	Video    VideoInfo     `db:"meta"`
}
