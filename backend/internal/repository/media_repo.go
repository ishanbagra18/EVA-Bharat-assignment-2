package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"media-sequencer-backend/internal/models"
)

type MediaRepository struct {
	db *sql.DB
}

func NewMediaRepository(db *sql.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

func (r *MediaRepository) GetAllMedia(ctx context.Context) ([]models.MediaItem, error) {
	query := `SELECT id, type, url, duration_seconds, created_at FROM media_items ORDER BY created_at ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query media items: %w", err)
	}
	defer rows.Close()

	var items []models.MediaItem
	for rows.Next() {
		var item models.MediaItem
		var url sql.NullString
		if err := rows.Scan(&item.ID, &item.Type, &url, &item.DurationSeconds, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan media item: %w", err)
		}
		if url.Valid {
			item.URL = &url.String
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *MediaRepository) GetMediaByID(ctx context.Context, id string) (*models.MediaItem, error) {
	query := `SELECT id, type, url, duration_seconds, created_at FROM media_items WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var item models.MediaItem
	var url sql.NullString
	if err := row.Scan(&item.ID, &item.Type, &url, &item.DurationSeconds, &item.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan media item by id: %w", err)
	}
	if url.Valid {
		item.URL = &url.String
	}
	return &item, nil
}

func (r *MediaRepository) CreateMedia(ctx context.Context, item models.MediaItem) (*models.MediaItem, error) {
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now().UTC()
	}

	query := `INSERT INTO media_items (id, type, url, duration_seconds, created_at) VALUES (?, ?, ?, ?, ?)`
	var urlVal *string
	if item.URL != nil && *item.URL != "" {
		urlVal = item.URL
	}

	_, err := r.db.ExecContext(ctx, query, item.ID, item.Type, urlVal, item.DurationSeconds, item.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert media item: %w", err)
	}

	return &item, nil
}
