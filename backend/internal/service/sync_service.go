package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"media-sequencer-backend/internal/broadcaster"
	"media-sequencer-backend/internal/models"
	"media-sequencer-backend/internal/repository"
)

type SyncService struct {
	mu          sync.RWMutex
	state       models.SyncState
	broadcaster *broadcaster.Broadcaster
	mediaRepo   *repository.MediaRepository
}

func NewSyncService(b *broadcaster.Broadcaster, mediaRepo *repository.MediaRepository) *SyncService {
	return &SyncService{
		broadcaster: b,
		mediaRepo:   mediaRepo,
	}
}

func (s *SyncService) GetStatus(ctx context.Context) models.SyncState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now().UTC()
	if s.state.Active && now.After(s.state.EndsAt) {
		// Auto expire sync state
		s.state.Active = false
	}
	return s.state
}

func (s *SyncService) TriggerSync(ctx context.Context, mediaID string, durationSec int) (*models.SyncState, error) {
	media, err := s.mediaRepo.GetMediaByID(ctx, mediaID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sync media: %w", err)
	}
	if media == nil {
		return nil, fmt.Errorf("media item %s not found", mediaID)
	}

	if durationSec <= 0 {
		durationSec = media.DurationSeconds
	}

	now := time.Now().UTC()
	endsAt := now.Add(time.Duration(durationSec) * time.Second)

	s.mu.Lock()
	s.state = models.SyncState{
		Active:      true,
		MediaID:     mediaID,
		Media:       media,
		StartedAt:   now,
		DurationSec: durationSec,
		EndsAt:      endsAt,
	}
	currentState := s.state
	s.mu.Unlock()

	// Broadcast SSE sync event
	payloadBytes, err := json.Marshal(currentState)
	if err == nil {
		s.broadcaster.Publish("sync", string(payloadBytes))
	}

	return &currentState, nil
}
