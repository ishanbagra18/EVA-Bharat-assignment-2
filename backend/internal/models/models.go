package models

import "time"

type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
	MediaTypeBlank MediaType = "blank"
)

type MediaItem struct {
	ID              string    `json:"id"`
	Type            MediaType `json:"type"`
	URL             *string   `json:"url,omitempty"`
	DurationSeconds int       `json:"duration_seconds"`
	CreatedAt       time.Time `json:"created_at"`
}

type PlaylistEntry struct {
	ID         string    `json:"id"`
	WindowID   string    `json:"window_id"`
	MediaID    string    `json:"media_id"`
	OrderIndex int       `json:"order_index"`
	Media      MediaItem `json:"media"`
}

type Window struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	CycleLengthSeconds int             `json:"cycle_length_seconds"`
	CycleEpoch         time.Time       `json:"cycle_epoch"`
	Playlist           []PlaylistEntry `json:"playlist"`
}

type SyncState struct {
	Active      bool       `json:"active"`
	MediaID     string     `json:"media_id,omitempty"`
	Media       *MediaItem `json:"media,omitempty"`
	StartedAt   time.Time  `json:"started_at"`
	DurationSec int        `json:"duration_seconds"`
	EndsAt      time.Time  `json:"ends_at"`
}

type CurrentPlaybackResponse struct {
	WindowID         string     `json:"window_id"`
	Media            *MediaItem `json:"media"`
	OffsetSeconds    float64    `json:"offset_seconds"`
	IsSync           bool       `json:"is_sync"`
	SyncRemainingSec float64    `json:"sync_remaining_sec,omitempty"`
	Message          string     `json:"message,omitempty"`
}

type AddMediaRequest struct {
	MediaID         string    `json:"media_id,omitempty"`
	Type            MediaType `json:"type,omitempty"`
	URL             *string   `json:"url,omitempty"`
	DurationSeconds int       `json:"duration_seconds,omitempty"`
}

type TriggerSyncRequest struct {
	MediaID         string `json:"media_id"`
	DurationSeconds int    `json:"duration_seconds"`
}
