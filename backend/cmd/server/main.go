package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"media-sequencer-backend/internal/broadcaster"
	"media-sequencer-backend/internal/db"
	"media-sequencer-backend/internal/handlers"
	"media-sequencer-backend/internal/repository"
	"media-sequencer-backend/internal/service"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./sequencer.db"
	}

	// 1. Initialize SQLite Database & Run Schema Migrations / Seeding
	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// 2. Initialize Repositories
	windowRepo := repository.NewWindowRepository(database)
	mediaRepo := repository.NewMediaRepository(database)

	// 3. Initialize Broadcaster & Services
	b := broadcaster.NewBroadcaster()
	playbackSvc := service.NewPlaybackService()
	syncSvc := service.NewSyncService(b, mediaRepo)

	// 4. Initialize HTTP Server Handlers
	srv := handlers.NewServer(windowRepo, mediaRepo, playbackSvc, syncSvc, b)

	// 5. Setup Chi Router & Middleware
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS Configuration for frontend clients
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// 6. Register API Endpoints
	r.Get("/health", srv.HealthHandler)
	r.Get("/events", srv.EventsHandler)

	r.Get("/windows", srv.GetWindows)
	r.Post("/windows", srv.CreateWindow)
	r.Get("/windows/{id}", srv.GetWindow)
	r.Get("/windows/{id}/now", srv.GetWindowNow)
	r.Post("/windows/{id}/media", srv.AddWindowMedia)

	r.Get("/media", srv.GetMedia)
	r.Post("/media", srv.CreateMedia)

	r.Post("/sync", srv.TriggerSync)
	r.Get("/sync/status", srv.GetSyncStatus)

	log.Printf("🚀 Media Sequencer Backend listening on http://0.0.0.0:%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}
