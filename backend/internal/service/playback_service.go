package service

import (
	"errors"
	"math"
	"time"

	"media-sequencer-backend/internal/models"
)

var ErrEmptyPlaylist = errors.New("playlist is empty")

type PlaybackService struct{}

func NewPlaybackService() *PlaybackService {
	return &PlaybackService{}
}

// CurrentItem computes what should be playing right now for a given window & playlist.
// Pure, deterministic calculation based on reference epoch and total playlist duration.
func (s *PlaybackService) CurrentItem(window models.Window, playlist []models.PlaylistEntry, now time.Time) (models.MediaItem, float64, error) {
	if len(playlist) == 0 {
		return models.MediaItem{}, 0, ErrEmptyPlaylist
	}

	playlistTotal := 0.0
	for _, entry := range playlist {
		if entry.Media.DurationSeconds > 0 {
			playlistTotal += float64(entry.Media.DurationSeconds)
		}
	}

	if playlistTotal <= 0 {
		return models.MediaItem{}, 0, ErrEmptyPlaylist
	}

	cycleLen := float64(window.CycleLengthSeconds)
	if cycleLen <= 0 {
		cycleLen = 18000
	}

	elapsedSec := now.Sub(window.CycleEpoch).Seconds()
	if elapsedSec < 0 {
		elapsedSec = math.Mod(elapsedSec, cycleLen)
		if elapsedSec < 0 {
			elapsedSec += cycleLen
		}
	}

	elapsedInCycle := math.Mod(elapsedSec, cycleLen)
	positionInLoop := math.Mod(elapsedInCycle, playlistTotal)

	cursor := 0.0
	for _, entry := range playlist {
		d := float64(entry.Media.DurationSeconds)
		if d <= 0 {
			continue
		}
		if positionInLoop < cursor+d {
			offset := positionInLoop - cursor
			return entry.Media, offset, nil
		}
		cursor += d
	}

	last := playlist[len(playlist)-1]
	return last.Media, float64(last.Media.DurationSeconds), nil
}
