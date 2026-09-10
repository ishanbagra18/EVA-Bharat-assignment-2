package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"media-sequencer-backend/internal/models"
)

type WindowRepository struct {
	db *sql.DB
}

func NewWindowRepository(db *sql.DB) *WindowRepository {
	return &WindowRepository{db: db}
}

func (r *WindowRepository) GetAllWindows(ctx context.Context) ([]models.Window, error) {
	query := `SELECT id, name, cycle_length_seconds, cycle_epoch FROM windows ORDER BY name ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query windows: %w", err)
	}
	defer rows.Close()

	var windows []models.Window
	for rows.Next() {
		var win models.Window
		if err := rows.Scan(&win.ID, &win.Name, &win.CycleLengthSeconds, &win.CycleEpoch); err != nil {
			return nil, fmt.Errorf("failed to scan window: %w", err)
		}
		win.Playlist = make([]models.PlaylistEntry, 0)
		windows = append(windows, win)
	}

	for i := range windows {
		playlist, err := r.GetWindowPlaylist(ctx, windows[i].ID)
		if err != nil {
			return nil, err
		}
		windows[i].Playlist = playlist
	}

	return windows, nil
}

func (r *WindowRepository) GetWindowByID(ctx context.Context, id string) (*models.Window, error) {
	query := `SELECT id, name, cycle_length_seconds, cycle_epoch FROM windows WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var win models.Window
	if err := row.Scan(&win.ID, &win.Name, &win.CycleLengthSeconds, &win.CycleEpoch); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan window by id: %w", err)
	}

	playlist, err := r.GetWindowPlaylist(ctx, win.ID)
	if err != nil {
		return nil, err
	}
	win.Playlist = playlist
	return &win, nil
}

func (r *WindowRepository) GetWindowPlaylist(ctx context.Context, windowID string) ([]models.PlaylistEntry, error) {
	query := `
		SELECT 
			e.id, e.window_id, e.media_id, e.order_index,
			m.id, m.type, m.url, m.duration_seconds, m.created_at
		FROM window_playlist_entries e
		JOIN media_items m ON e.media_id = m.id
		WHERE e.window_id = ?
		ORDER BY e.order_index ASC
	`
	rows, err := r.db.QueryContext(ctx, query, windowID)
	if err != nil {
		return nil, fmt.Errorf("failed to query playlist entries: %w", err)
	}
	defer rows.Close()

	entries := make([]models.PlaylistEntry, 0)
	for rows.Next() {
		var entry models.PlaylistEntry
		var m models.MediaItem
		var url sql.NullString

		if err := rows.Scan(
			&entry.ID, &entry.WindowID, &entry.MediaID, &entry.OrderIndex,
			&m.ID, &m.Type, &url, &m.DurationSeconds, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan playlist entry: %w", err)
		}

		if url.Valid {
			m.URL = &url.String
		}
		entry.Media = m
		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *WindowRepository) CreateWindow(ctx context.Context, win models.Window) (*models.Window, error) {
	query := `INSERT INTO windows (id, name, cycle_length_seconds, cycle_epoch) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, win.ID, win.Name, win.CycleLengthSeconds, win.CycleEpoch)
	if err != nil {
		return nil, fmt.Errorf("failed to create window: %w", err)
	}
	win.Playlist = make([]models.PlaylistEntry, 0)
	return &win, nil
}

func (r *WindowRepository) AddMediaToPlaylist(ctx context.Context, windowID, mediaID string, targetOrderIndex int) ([]models.PlaylistEntry, error) {
	if targetOrderIndex < 0 {
		var maxIndex sql.NullInt64
		err := r.db.QueryRowContext(ctx, "SELECT MAX(order_index) FROM window_playlist_entries WHERE window_id = ?", windowID).Scan(&maxIndex)
		if err != nil {
			return nil, fmt.Errorf("failed to query max order_index: %w", err)
		}
		if maxIndex.Valid {
			targetOrderIndex = int(maxIndex.Int64) + 1
		} else {
			targetOrderIndex = 0
		}
	}

	entryID := uuid.New().String()
	query := `INSERT INTO window_playlist_entries (id, window_id, media_id, order_index) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, entryID, windowID, mediaID, targetOrderIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to insert playlist entry: %w", err)
	}

	return r.GetWindowPlaylist(ctx, windowID)
}
