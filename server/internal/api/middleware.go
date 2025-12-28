package api

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/satetsu888/agentrace/server/internal/config"
)

type Middleware struct {
	cfg *config.Config
}

func NewMiddleware(cfg *config.Config) *Middleware {
	return &Middleware{cfg: cfg}
}

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get Bearer token from Authorization header
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

		// For Step 1, validate against fixed API key
		if m.cfg.APIKeyFixed == "" {
			http.Error(w, `{"error": "server api key not configured"}`, http.StatusInternalServerError)
			return
		}

		if token != m.cfg.APIKeyFixed {
			http.Error(w, `{"error": "invalid api key"}`, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

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
