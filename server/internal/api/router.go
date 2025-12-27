package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/satetsu888/agentrace/server/internal/config"
	"github.com/satetsu888/agentrace/server/internal/repository"
)

func NewRouter(cfg *config.Config, repos *repository.Repositories) http.Handler {
	r := mux.NewRouter()

	// Middleware
	authMiddleware := NewAuthMiddleware(cfg)

	// Handlers
	ingestHandler := NewIngestHandler(repos)
	sessionHandler := NewSessionHandler(repos)

	// API routes (authenticated)
	api := r.PathPrefix("/api").Subrouter()
	api.Use(authMiddleware.Authenticate)

	api.HandleFunc("/ingest", ingestHandler.Handle).Methods("POST")
	api.HandleFunc("/sessions", sessionHandler.List).Methods("GET")
	api.HandleFunc("/sessions/{id}", sessionHandler.Get).Methods("GET")

	// Health check (no auth)
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	}).Methods("GET")

	return r
}
