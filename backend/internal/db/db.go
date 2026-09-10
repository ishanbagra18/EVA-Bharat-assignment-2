package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// SQLite pragma optimizations
	if _, err := db.Exec(`
		PRAGMA foreign_keys = ON;
		PRAGMA journal_mode = WAL;
	`); err != nil {
		return nil, fmt.Errorf("failed to set sqlite pragmas: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	if err := seedData(db); err != nil {
		return nil, fmt.Errorf("failed to seed initial data: %w", err)
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS windows (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		cycle_length_seconds INTEGER NOT NULL DEFAULT 18000,
		cycle_epoch DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS media_items (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL CHECK (type IN ('image','video','blank')),
		url TEXT,
		duration_seconds INTEGER NOT NULL CHECK (duration_seconds > 0),
		created_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS window_playlist_entries (
		id TEXT PRIMARY KEY,
		window_id TEXT NOT NULL REFERENCES windows(id) ON DELETE CASCADE,
		media_id TEXT NOT NULL REFERENCES media_items(id) ON DELETE CASCADE,
		order_index INTEGER NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_wpe_window ON window_playlist_entries(window_id, order_index);
	`
	_, err := db.Exec(schema)
	return err
}

func seedData(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM windows").Scan(&count)
	if err != nil {
		return err
	}

	// Only seed if empty
	if count > 0 {
		return nil
	}

	log.Println("Seeding initial windows and media items into SQLite...")

	// Fixed epoch starting midnight 2026-01-01 UTC for consistency
	epoch := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Now().UTC()

	// Seed Windows
	windows := []struct {
		id   string
		name string
	}{
		{"w1", "Window 1 — Main Screen"},
		{"w2", "Window 2 — Display Wall"},
	}

	for _, w := range windows {
		_, err := db.Exec(
			"INSERT INTO windows (id, name, cycle_length_seconds, cycle_epoch) VALUES (?, ?, ?, ?)",
			w.id, w.name, 18000, epoch,
		)
		if err != nil {
			return fmt.Errorf("failed to seed window %s: %w", w.id, err)
		}
	}

	// Seed Media Items
	url1 := "https://images.unsplash.com/photo-1579546929518-9e396f3cc809?w=1200"
	url2 := "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4"
	url4 := "https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?w=1200"

	mediaItems := []struct {
		id       string
		mType    string
		url      *string
		duration int
	}{
		{"m1", "image", &url1, 15},
		{"m2", "video", &url2, 30},
		{"m3", "blank", nil, 10},
		{"m4", "image", &url4, 20},
	}

	for _, m := range mediaItems {
		_, err := db.Exec(
			"INSERT INTO media_items (id, type, url, duration_seconds, created_at) VALUES (?, ?, ?, ?, ?)",
			m.id, m.mType, m.url, m.duration, now,
		)
		if err != nil {
			return fmt.Errorf("failed to seed media item %s: %w", m.id, err)
		}
	}

	// Seed Playlists
	// w1 playlist: m1 (15s), m2 (30s), m4 (20s), m2 (30s)
	w1Entries := []string{"m1", "m2", "m4", "m2"}
	for idx, mID := range w1Entries {
		id := uuid.New().String()
		_, err := db.Exec(
			"INSERT INTO window_playlist_entries (id, window_id, media_id, order_index) VALUES (?, ?, ?, ?)",
			id, "w1", mID, idx,
		)
		if err != nil {
			return fmt.Errorf("failed to seed w1 entry %d: %w", idx, err)
		}
	}

	// w2 playlist: m2 (30s), m3 (10s blank slot), m1 (15s)
	w2Entries := []string{"m2", "m3", "m1"}
	for idx, mID := range w2Entries {
		id := uuid.New().String()
		_, err := db.Exec(
			"INSERT INTO window_playlist_entries (id, window_id, media_id, order_index) VALUES (?, ?, ?, ?)",
			id, "w2", mID, idx,
		)
		if err != nil {
			return fmt.Errorf("failed to seed w2 entry %d: %w", idx, err)
		}
	}

	log.Println("SQLite seeding completed successfully.")
	return nil
}
