package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"media-sequencer-backend/internal/broadcaster"
	"media-sequencer-backend/internal/models"
	"media-sequencer-backend/internal/repository"
	"media-sequencer-backend/internal/service"
)

type Server struct {
	windowRepo  *repository.WindowRepository
	mediaRepo   *repository.MediaRepository
	playbackSvc *service.PlaybackService
	syncSvc     *service.SyncService
	broadcaster *broadcaster.Broadcaster
}

func NewServer(
	windowRepo *repository.WindowRepository,
	mediaRepo *repository.MediaRepository,
	playbackSvc *service.PlaybackService,
	syncSvc *service.SyncService,
	broadcaster *broadcaster.Broadcaster,
) *Server {
	return &Server{
		windowRepo:  windowRepo,
		mediaRepo:   mediaRepo,
		playbackSvc: playbackSvc,
		syncSvc:     syncSvc,
		broadcaster: broadcaster,
	}
}

func jsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, statusCode int, message string) {
	jsonResponse(w, statusCode, map[string]string{"error": message})
}

// GET /health
func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
	})
}

// GET /windows
func (s *Server) GetWindows(w http.ResponseWriter, r *http.Request) {
	windows, err := s.windowRepo.GetAllWindows(r.Context())
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, windows)
}

// GET /windows/{id}
func (s *Server) GetWindow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	win, err := s.windowRepo.GetWindowByID(r.Context(), id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if win == nil {
		jsonError(w, http.StatusNotFound, "window not found")
		return
	}
	jsonResponse(w, http.StatusOK, win)
}

// GET /windows/{id}/now
func (s *Server) GetWindowNow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	win, err := s.windowRepo.GetWindowByID(r.Context(), id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if win == nil {
		jsonError(w, http.StatusNotFound, "window not found")
		return
	}

	now := time.Now().UTC()
	syncState := s.syncSvc.GetStatus(r.Context())

	// Check if sync is currently active
	if syncState.Active && now.Before(syncState.EndsAt) {
		offset := now.Sub(syncState.StartedAt).Seconds()
		remaining := syncState.EndsAt.Sub(now).Seconds()
		jsonResponse(w, http.StatusOK, models.CurrentPlaybackResponse{
			WindowID:         win.ID,
			Media:            syncState.Media,
			OffsetSeconds:    offset,
			IsSync:           true,
			SyncRemainingSec: remaining,
			Message:          "Global broadcast sync in progress",
		})
		return
	}

	// Calculate deterministic virtual clock playback
	item, offset, err := s.playbackSvc.CurrentItem(*win, win.Playlist, now)
	if err != nil {
		jsonResponse(w, http.StatusOK, models.CurrentPlaybackResponse{
			WindowID: win.ID,
			Media:    nil,
			IsSync:   false,
			Message:  "Playlist is empty or invalid",
		})
		return
	}

	jsonResponse(w, http.StatusOK, models.CurrentPlaybackResponse{
		WindowID:      win.ID,
		Media:         &item,
		OffsetSeconds: offset,
		IsSync:        false,
	})
}

// POST /windows
func (s *Server) CreateWindow(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name               string `json:"name"`
		CycleLengthSeconds int    `json:"cycle_length_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		jsonError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.CycleLengthSeconds <= 0 {
		req.CycleLengthSeconds = 18000 // 5 hours default
	}

	newWin := models.Window{
		ID:                 uuid.New().String(),
		Name:               req.Name,
		CycleLengthSeconds: req.CycleLengthSeconds,
		CycleEpoch:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	created, err := s.windowRepo.CreateWindow(r.Context(), newWin)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, created)
}

// POST /windows/{id}/media
func (s *Server) AddWindowMedia(w http.ResponseWriter, r *http.Request) {
	windowID := chi.URLParam(r, "id")
	win, err := s.windowRepo.GetWindowByID(r.Context(), windowID)
	if err != nil || win == nil {
		jsonError(w, http.StatusNotFound, "window not found")
		return
	}

	var req models.AddMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	mediaID := req.MediaID

	// Support inline creation if media_id is not provided
	if mediaID == "" {
		if req.Type == "" {
			jsonError(w, http.StatusBadRequest, "media_id or type is required")
			return
		}
		if req.Type != models.MediaTypeBlank && (req.URL == nil || *req.URL == "") {
			jsonError(w, http.StatusBadRequest, "url is required for image/video types")
			return
		}
		if req.DurationSeconds <= 0 {
			jsonError(w, http.StatusBadRequest, "duration_seconds must be greater than 0")
			return
		}

		newItem := models.MediaItem{
			ID:              uuid.New().String(),
			Type:            req.Type,
			URL:             req.URL,
			DurationSeconds: req.DurationSeconds,
			CreatedAt:       time.Now().UTC(),
		}

		createdItem, err := s.mediaRepo.CreateMedia(r.Context(), newItem)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create media item: %v", err))
			return
		}
		mediaID = createdItem.ID
	} else {
		// Verify media item exists
		existingMedia, err := s.mediaRepo.GetMediaByID(r.Context(), mediaID)
		if err != nil || existingMedia == nil {
			jsonError(w, http.StatusNotFound, fmt.Sprintf("media item %s not found", mediaID))
			return
		}
	}

	// Append to playlist at the end (-1)
	updatedPlaylist, err := s.windowRepo.AddMediaToPlaylist(r.Context(), windowID, mediaID, -1)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Broadcast SSE playlist_updated event
	payload, _ := json.Marshal(map[string]string{
		"window_id": windowID,
	})
	s.broadcaster.Publish("playlist_updated", string(payload))

	jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"window_id": windowID,
		"playlist":  updatedPlaylist,
	})
}

// GET /media
func (s *Server) GetMedia(w http.ResponseWriter, r *http.Request) {
	items, err := s.mediaRepo.GetAllMedia(r.Context())
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, items)
}

// POST /media
func (s *Server) CreateMedia(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type            models.MediaType `json:"type"`
		URL             *string          `json:"url"`
		DurationSeconds int              `json:"duration_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Type == "" {
		jsonError(w, http.StatusBadRequest, "type is required")
		return
	}
	if req.Type != models.MediaTypeBlank && (req.URL == nil || *req.URL == "") {
		jsonError(w, http.StatusBadRequest, "url is required for image or video media")
		return
	}
	if req.DurationSeconds <= 0 {
		jsonError(w, http.StatusBadRequest, "duration_seconds must be greater than 0")
		return
	}

	item := models.MediaItem{
		ID:              uuid.New().String(),
		Type:            req.Type,
		URL:             req.URL,
		DurationSeconds: req.DurationSeconds,
		CreatedAt:       time.Now().UTC(),
	}

	created, err := s.mediaRepo.CreateMedia(r.Context(), item)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, created)
}

// POST /sync
func (s *Server) TriggerSync(w http.ResponseWriter, r *http.Request) {
	var req models.TriggerSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.MediaID == "" {
		jsonError(w, http.StatusBadRequest, "media_id is required")
		return
	}

	syncState, err := s.syncSvc.TriggerSync(r.Context(), req.MediaID, req.DurationSeconds)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, syncState)
}

// GET /sync/status
func (s *Server) GetSyncStatus(w http.ResponseWriter, r *http.Request) {
	syncState := s.syncSvc.GetStatus(r.Context())
	jsonResponse(w, http.StatusOK, syncState)
}

// GET /events
func (s *Server) EventsHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := s.broadcaster.Subscribe()
	defer s.broadcaster.Unsubscribe(ch)

	// Send initial ping to confirm SSE connection established
	fmt.Fprint(w, "event: connected\ndata: {\"status\":\"connected\"}\n\n")
	flusher.Flush()

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprint(w, msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
