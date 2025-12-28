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
	mw := NewMiddleware(cfg, repos)

	// Apply request logger to all routes
	r.Use(mw.RequestLogger)

	// Handlers
	ingestHandler := NewIngestHandler(repos)
	sessionHandler := NewSessionHandler(repos)
	authHandler := NewAuthHandler(cfg, repos)

	// Auth routes (no auth required)
	r.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	r.HandleFunc("/auth/session", authHandler.Session).Methods("GET")

	// API routes (Bearer auth - for CLI)
	apiBearer := r.PathPrefix("/api").Subrouter()
	apiBearer.Use(mw.AuthenticateBearer)
	apiBearer.HandleFunc("/ingest", ingestHandler.Handle).Methods("POST")
	apiBearer.HandleFunc("/auth/web-session", authHandler.CreateWebSession).Methods("POST")

	// API routes (Session auth - for Web)
	apiSession := r.PathPrefix("/api").Subrouter()
	apiSession.Use(mw.AuthenticateSession)
	apiSession.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST")
	apiSession.HandleFunc("/me", authHandler.Me).Methods("GET")
	apiSession.HandleFunc("/users", authHandler.ListUsers).Methods("GET")
	apiSession.HandleFunc("/keys", authHandler.ListKeys).Methods("GET")
	apiSession.HandleFunc("/keys", authHandler.CreateKey).Methods("POST")
	apiSession.HandleFunc("/keys/{id}", authHandler.DeleteKey).Methods("DELETE")

	// API routes (Bearer or Session auth - for both)
	apiBoth := r.PathPrefix("/api").Subrouter()
	apiBoth.Use(mw.AuthenticateBearerOrSession)
	apiBoth.HandleFunc("/sessions", sessionHandler.List).Methods("GET")
	apiBoth.HandleFunc("/sessions/{id}", sessionHandler.Get).Methods("GET")

	// Health check (no auth)
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	}).Methods("GET")

	return r
}
