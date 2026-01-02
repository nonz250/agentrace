package api

import (
	"encoding/json"
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
	planDocumentHandler := NewPlanDocumentHandler(repos)
	projectHandler := NewProjectHandler(repos)

	// Auth routes (no auth required)
	r.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	r.HandleFunc("/auth/login/apikey", authHandler.LoginWithAPIKey).Methods("POST")
	r.HandleFunc("/auth/session", authHandler.Session).Methods("GET")
	r.HandleFunc("/auth/github", authHandler.GitHubAuth).Methods("GET")
	r.HandleFunc("/auth/github/callback", authHandler.GitHubCallback).Methods("GET")

	// API routes (Bearer auth - for CLI)
	apiBearer := r.PathPrefix("/api").Subrouter()
	apiBearer.Use(mw.AuthenticateBearer)
	apiBearer.HandleFunc("/ingest", ingestHandler.Handle).Methods("POST")
	apiBearer.HandleFunc("/auth/web-session", authHandler.CreateWebSession).Methods("POST")

	// API routes (Bearer or Session auth - for CLI and Web)
	apiBearerOrSession := r.PathPrefix("/api").Subrouter()
	apiBearerOrSession.Use(mw.AuthenticateBearerOrSession)
	apiBearerOrSession.HandleFunc("/plans", planDocumentHandler.Create).Methods("POST")
	apiBearerOrSession.HandleFunc("/plans/{id}", planDocumentHandler.Update).Methods("PATCH")
	apiBearerOrSession.HandleFunc("/plans/{id}", planDocumentHandler.Delete).Methods("DELETE")
	apiBearerOrSession.HandleFunc("/plans/{id}/status", planDocumentHandler.SetStatus).Methods("PATCH")

	// API routes (Session auth - for Web)
	apiSession := r.PathPrefix("/api").Subrouter()
	apiSession.Use(mw.AuthenticateSession)
	apiSession.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST")
	apiSession.HandleFunc("/me", authHandler.Me).Methods("GET")
	apiSession.HandleFunc("/me", authHandler.UpdateMe).Methods("PATCH")
	apiSession.HandleFunc("/keys", authHandler.ListKeys).Methods("GET")
	apiSession.HandleFunc("/keys", authHandler.CreateKey).Methods("POST")
	apiSession.HandleFunc("/keys/{id}", authHandler.DeleteKey).Methods("DELETE")

	// API routes (Optional auth - public read access)
	apiOptional := r.PathPrefix("/api").Subrouter()
	apiOptional.Use(mw.OptionalBearerOrSession)
	apiOptional.HandleFunc("/sessions", sessionHandler.List).Methods("GET")
	apiOptional.HandleFunc("/sessions/{id}", sessionHandler.Get).Methods("GET")
	apiOptional.HandleFunc("/plans", planDocumentHandler.List).Methods("GET")
	apiOptional.HandleFunc("/plans/{id}", planDocumentHandler.Get).Methods("GET")
	apiOptional.HandleFunc("/plans/{id}/events", planDocumentHandler.GetEvents).Methods("GET")
	apiOptional.HandleFunc("/projects", projectHandler.List).Methods("GET")
	apiOptional.HandleFunc("/projects/{id}", projectHandler.Get).Methods("GET")
	apiOptional.HandleFunc("/users", authHandler.ListUsers).Methods("GET")

	// Health check (no auth)
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	}).Methods("GET")

	// Auth config (no auth) - for frontend to check available OAuth providers
	r.HandleFunc("/auth/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := struct {
			GitHubEnabled bool `json:"github_enabled"`
		}{
			GitHubEnabled: cfg.IsGitHubOAuthEnabled(),
		}
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")

	// Setup redirect (for CLI init flow)
	// If WEB_URL is set, redirect to frontend; otherwise assume same origin
	r.HandleFunc("/setup", func(w http.ResponseWriter, r *http.Request) {
		if cfg.WebURL != "" {
			// Redirect to frontend with query params preserved
			redirectURL := cfg.WebURL + "/setup"
			if r.URL.RawQuery != "" {
				redirectURL += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, redirectURL, http.StatusFound)
			return
		}
		// If no WEB_URL, return 404 (frontend should be served from same origin)
		http.NotFound(w, r)
	}).Methods("GET")

	return r
}
