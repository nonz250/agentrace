package api

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/satetsu888/agentrace/server/internal/config"
	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const (
	userContextKey   contextKey = "user"
	userIDContextKey contextKey = "userID"
)

type Middleware struct {
	cfg   *config.Config
	repos *repository.Repositories
}

func NewMiddleware(cfg *config.Config, repos *repository.Repositories) *Middleware {
	return &Middleware{cfg: cfg, repos: repos}
}

// GetUserFromContext returns the authenticated user from context
func GetUserFromContext(ctx context.Context) *domain.User {
	user, _ := ctx.Value(userContextKey).(*domain.User)
	return user
}

// GetUserIDFromContext returns the authenticated user ID from context
func GetUserIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(userIDContextKey).(string)
	return userID
}

// setUserContext sets the user in the request context
func setUserContext(r *http.Request, user *domain.User) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, userContextKey, user)
	ctx = context.WithValue(ctx, userIDContextKey, user.ID)
	return r.WithContext(ctx)
}

// AuthenticateBearer validates Bearer token (API key) authentication
func (m *Middleware) AuthenticateBearer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"error": "invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Check fixed API key for backward compatibility (dev mode)
		if m.cfg.APIKeyFixed != "" && token == m.cfg.APIKeyFixed {
			next.ServeHTTP(w, r)
			return
		}

		// Validate API key from database
		ctx := r.Context()

		// Find API key by iterating (bcrypt comparison is slow, so we'll use prefix matching first)
		// For production, we'd want to store a hash that can be looked up directly
		keys, err := m.findAPIKeyByToken(ctx, token)
		if err != nil || keys == nil {
			http.Error(w, `{"error": "invalid api key"}`, http.StatusUnauthorized)
			return
		}

		// Update last used at
		_ = m.repos.APIKey.UpdateLastUsedAt(ctx, keys.ID)

		// Get user
		user, err := m.repos.User.FindByID(ctx, keys.UserID)
		if err != nil || user == nil {
			http.Error(w, `{"error": "user not found"}`, http.StatusUnauthorized)
			return
		}

		// Set user in context
		r = setUserContext(r, user)
		next.ServeHTTP(w, r)
	})
}

// findAPIKeyByToken finds an API key by comparing bcrypt hash
func (m *Middleware) findAPIKeyByToken(ctx context.Context, token string) (*domain.APIKey, error) {
	// Get all API keys and compare
	// This is not efficient, but works for memory storage
	// For production with PostgreSQL, we'd use a different approach
	users, err := m.repos.User.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		keys, err := m.repos.APIKey.FindByUserID(ctx, user.ID)
		if err != nil {
			continue
		}
		for _, key := range keys {
			if err := bcrypt.CompareHashAndPassword([]byte(key.KeyHash), []byte(token)); err == nil {
				return key, nil
			}
		}
	}
	return nil, nil
}

// AuthenticateSession validates session cookie authentication
func (m *Middleware) AuthenticateSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, `{"error": "missing session cookie"}`, http.StatusUnauthorized)
			return
		}

		ctx := r.Context()

		// Find session by token
		session, err := m.repos.WebSession.FindByToken(ctx, cookie.Value)
		if err != nil || session == nil {
			http.Error(w, `{"error": "invalid session"}`, http.StatusUnauthorized)
			return
		}

		// Check if session is expired
		if session.IsExpired() {
			_ = m.repos.WebSession.Delete(ctx, session.ID)
			http.Error(w, `{"error": "session expired"}`, http.StatusUnauthorized)
			return
		}

		// Get user
		user, err := m.repos.User.FindByID(ctx, session.UserID)
		if err != nil || user == nil {
			http.Error(w, `{"error": "user not found"}`, http.StatusUnauthorized)
			return
		}

		// Set user in context
		r = setUserContext(r, user)
		next.ServeHTTP(w, r)
	})
}

// AuthenticateBearerOrSession validates either Bearer token or session cookie
func (m *Middleware) AuthenticateBearerOrSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try Bearer first
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			m.AuthenticateBearer(next).ServeHTTP(w, r)
			return
		}

		// Try session cookie
		_, err := r.Cookie("session")
		if err == nil {
			m.AuthenticateSession(next).ServeHTTP(w, r)
			return
		}

		http.Error(w, `{"error": "missing authentication"}`, http.StatusUnauthorized)
	})
}

// RequestLogger logs requests in dev mode
func (m *Middleware) RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.cfg.IsDevMode() {
			next.ServeHTTP(w, r)
			return
		}

		// Read and log request body
		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		log.Printf("[DEBUG] %s %s", r.Method, r.URL.Path)
		if len(bodyBytes) > 0 {
			log.Printf("[DEBUG] Body: %s", string(bodyBytes))
		}

		next.ServeHTTP(w, r)
	})
}
