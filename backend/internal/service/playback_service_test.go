package service

import (
	"testing"
	"time"

	"media-sequencer-backend/internal/models"
)

func TestPlaybackService_CurrentItem(t *testing.T) {
	svc := NewPlaybackService()
	epoch := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	win := models.Window{
		ID:                 "w1",
		Name:               "Test Window",
		CycleLengthSeconds: 18000, // 5 hours
		CycleEpoch:         epoch,
	}

	m1 := models.MediaItem{ID: "m1", Type: models.MediaTypeImage, DurationSeconds: 15}
	m2 := models.MediaItem{ID: "m2", Type: models.MediaTypeVideo, DurationSeconds: 30}
	m3 := models.MediaItem{ID: "m3", Type: models.MediaTypeBlank, DurationSeconds: 10}

	playlist := []models.PlaylistEntry{
		{ID: "pe1", WindowID: "w1", MediaID: "m1", OrderIndex: 0, Media: m1},
		{ID: "pe2", WindowID: "w1", MediaID: "m2", OrderIndex: 1, Media: m2},
		{ID: "pe3", WindowID: "w1", MediaID: "m3", OrderIndex: 2, Media: m3},
	}
	// Total playlist duration = 15 + 30 + 10 = 55s

	tests := []struct {
		name           string
		now            time.Time
		expectedMediaID string
		expectedOffset float64
	}{
		{
			name:           "Exact Epoch (t=0s)",
			now:            epoch,
			expectedMediaID: "m1",
			expectedOffset: 0.0,
		},
		{
			name:           "Midway through m1 (t=10s)",
			now:            epoch.Add(10 * time.Second),
			expectedMediaID: "m1",
			expectedOffset: 10.0,
		},
		{
			name:           "Start of m2 (t=15s)",
			now:            epoch.Add(15 * time.Second),
			expectedMediaID: "m2",
			expectedOffset: 0.0,
		},
		{
			name:           "Midway through m2 (t=30s)",
			now:            epoch.Add(30 * time.Second),
			expectedMediaID: "m2",
			expectedOffset: 15.0,
		},
		{
			name:           "Start of m3 blank slot (t=45s)",
			now:            epoch.Add(45 * time.Second),
			expectedMediaID: "m3",
			expectedOffset: 0.0,
		},
		{
			name:           "Loop wrap 1 (t=55s -> returns to m1 at offset 0)",
			now:            epoch.Add(55 * time.Second),
			expectedMediaID: "m1",
			expectedOffset: 0.0,
		},
		{
			name:           "Loop wrap 1 mid m2 (t=70s -> 70%55=15 -> start of m2)",
			now:            epoch.Add(70 * time.Second),
			expectedMediaID: "m2",
			expectedOffset: 0.0,
		},
		{
			name:           "5-hour Cycle Wrap (t=18005s -> 18005%18000=5s -> m1 at offset 5)",
			now:            epoch.Add(18005 * time.Second),
			expectedMediaID: "m1",
			expectedOffset: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, offset, err := svc.CurrentItem(win, playlist, tt.now)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if item.ID != tt.expectedMediaID {
				t.Errorf("expected media %s, got %s", tt.expectedMediaID, item.ID)
			}
			if offset != tt.expectedOffset {
				t.Errorf("expected offset %.2f, got %.2f", tt.expectedOffset, offset)
			}
		})
	}
}

func TestPlaybackService_EmptyPlaylist(t *testing.T) {
	svc := NewPlaybackService()
	epoch := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	win := models.Window{ID: "w1", CycleLengthSeconds: 18000, CycleEpoch: epoch}

	_, _, err := svc.CurrentItem(win, []models.PlaylistEntry{}, epoch)
	if err != ErrEmptyPlaylist {
		t.Errorf("expected ErrEmptyPlaylist, got %v", err)
	}
}
