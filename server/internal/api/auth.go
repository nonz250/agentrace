package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/satetsu888/agentrace/server/internal/config"
	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const (
	apiKeyPrefix       = "agtr_"
	apiKeyLength       = 32
	sessionTokenLength = 32
	sessionDuration    = 7 * 24 * time.Hour   // 7 days
	webSessionDuration = 10 * time.Minute     // 10 minutes for CLI login
)

type AuthHandler struct {
	cfg   *config.Config
	repos *repository.Repositories
}

func NewAuthHandler(cfg *config.Config, repos *repository.Repositories) *AuthHandler {
	return &AuthHandler{cfg: cfg, repos: repos}
}

// RegisterRequest is the request body for user registration
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResponse is the response for user registration
type RegisterResponse struct {
	User   *domain.User `json:"user"`
	APIKey string       `json:"api_key"`
}

// LoginRequest is the request body for login with email/password
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginWithAPIKeyRequest is the request body for login with API key
type LoginWithAPIKeyRequest struct {
	APIKey string `json:"api_key"`
}

// LoginResponse is the response for login
type LoginResponse struct {
	User *domain.User `json:"user"`
}

// WebSessionResponse is the response for web session creation
type WebSessionResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// CreateKeyRequest is the request body for creating an API key
type CreateKeyRequest struct {
	Name string `json:"name"`
}

// CreateKeyResponse is the response for creating an API key
type CreateKeyResponse struct {
	Key    *APIKeyInfo `json:"key"`
	APIKey string      `json:"api_key"`
}

// APIKeyInfo is the public info of an API key
type APIKeyInfo struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ListKeysResponse is the response for listing API keys
type ListKeysResponse struct {
	Keys []*APIKeyInfo `json:"keys"`
}

// MeResponse is the response for getting current user
type MeResponse struct {
	User *domain.User `json:"user"`
}

// UsersResponse is the response for listing users
type UsersResponse struct {
	Users []*domain.User `json:"users"`
}

// generateAPIKey generates a new API key
func generateAPIKey() (string, error) {
	bytes := make([]byte, apiKeyLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return apiKeyPrefix + base64.RawURLEncoding.EncodeToString(bytes), nil
}

// generateToken generates a random token
func generateToken() (string, error) {
	bytes := make([]byte, sessionTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// hashAPIKey hashes an API key using bcrypt
func hashAPIKey(key string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// checkPassword compares a password with a hash
func checkPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, `{"error": "email is required"}`, http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, `{"error": "password is required"}`, http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		http.Error(w, `{"error": "password must be at least 8 characters"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Check if email is already registered
	existingUser, err := h.repos.User.FindByEmail(ctx, req.Email)
	if err != nil {
		http.Error(w, `{"error": "failed to check email"}`, http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		http.Error(w, `{"error": "email already registered"}`, http.StatusConflict)
		return
	}

	// Hash password
	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		http.Error(w, `{"error": "failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	// Create user (DisplayName is empty, will show email)
	user := &domain.User{
		Email: req.Email,
	}
	if err := h.repos.User.Create(ctx, user); err != nil {
		http.Error(w, `{"error": "failed to create user"}`, http.StatusInternalServerError)
		return
	}

	// Create password credential
	passwordCred := &domain.PasswordCredential{
		UserID:       user.ID,
		PasswordHash: passwordHash,
	}
	if err := h.repos.PasswordCredential.Create(ctx, passwordCred); err != nil {
		http.Error(w, `{"error": "failed to create password credential"}`, http.StatusInternalServerError)
		return
	}

	// Generate API key
	rawKey, err := generateAPIKey()
	if err != nil {
		http.Error(w, `{"error": "failed to generate api key"}`, http.StatusInternalServerError)
		return
	}

	keyHash, err := hashAPIKey(rawKey)
	if err != nil {
		http.Error(w, `{"error": "failed to hash api key"}`, http.StatusInternalServerError)
		return
	}

	apiKey := &domain.APIKey{
		UserID:    user.ID,
		Name:      "Default",
		KeyHash:   keyHash,
		KeyPrefix: rawKey[:12] + "...",
	}
	if err := h.repos.APIKey.Create(ctx, apiKey); err != nil {
		http.Error(w, `{"error": "failed to create api key"}`, http.StatusInternalServerError)
		return
	}

	// Create web session for auto-login
	sessionToken, err := generateToken()
	if err != nil {
		http.Error(w, `{"error": "failed to generate session token"}`, http.StatusInternalServerError)
		return
	}

	webSession := &domain.WebSession{
		UserID:    user.ID,
		Token:     sessionToken,
		ExpiresAt: time.Now().Add(sessionDuration),
	}
	if err := h.repos.WebSession.Create(ctx, webSession); err != nil {
		http.Error(w, `{"error": "failed to create web session"}`, http.StatusInternalServerError)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionToken,
		Path:     "/",
		Expires:  webSession.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	resp := RegisterResponse{
		User:   user,
		APIKey: rawKey,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Login handles user login with email/password
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, `{"error": "email is required"}`, http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, `{"error": "password is required"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Find user by email
	user, err := h.repos.User.FindByEmail(ctx, req.Email)
	if err != nil {
		http.Error(w, `{"error": "failed to find user"}`, http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, `{"error": "invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	// Find password credential
	passwordCred, err := h.repos.PasswordCredential.FindByUserID(ctx, user.ID)
	if err != nil {
		http.Error(w, `{"error": "failed to find password credential"}`, http.StatusInternalServerError)
		return
	}
	if passwordCred == nil {
		http.Error(w, `{"error": "invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	// Check password
	if !checkPassword(req.Password, passwordCred.PasswordHash) {
		http.Error(w, `{"error": "invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	// Create web session
	sessionToken, err := generateToken()
	if err != nil {
		http.Error(w, `{"error": "failed to generate session token"}`, http.StatusInternalServerError)
		return
	}

	webSession := &domain.WebSession{
		UserID:    user.ID,
		Token:     sessionToken,
		ExpiresAt: time.Now().Add(sessionDuration),
	}
	if err := h.repos.WebSession.Create(ctx, webSession); err != nil {
		http.Error(w, `{"error": "failed to create web session"}`, http.StatusInternalServerError)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionToken,
		Path:     "/",
		Expires:  webSession.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	resp := LoginResponse{
		User: user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// LoginWithAPIKey handles user login with API key
func (h *AuthHandler) LoginWithAPIKey(w http.ResponseWriter, r *http.Request) {
	var req LoginWithAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.APIKey == "" {
		http.Error(w, `{"error": "api_key is required"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Find API key
	apiKey, user, err := h.findAPIKeyAndUser(ctx, req.APIKey)
	if err != nil || apiKey == nil || user == nil {
		http.Error(w, `{"error": "invalid api key"}`, http.StatusUnauthorized)
		return
	}

	// Update last used at
	_ = h.repos.APIKey.UpdateLastUsedAt(ctx, apiKey.ID)

	// Create web session
	sessionToken, err := generateToken()
	if err != nil {
		http.Error(w, `{"error": "failed to generate session token"}`, http.StatusInternalServerError)
		return
	}

	webSession := &domain.WebSession{
		UserID:    user.ID,
		Token:     sessionToken,
		ExpiresAt: time.Now().Add(sessionDuration),
	}
	if err := h.repos.WebSession.Create(ctx, webSession); err != nil {
		http.Error(w, `{"error": "failed to create web session"}`, http.StatusInternalServerError)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionToken,
		Path:     "/",
		Expires:  webSession.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	resp := LoginResponse{
		User: user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// findAPIKeyAndUser finds an API key and its associated user
func (h *AuthHandler) findAPIKeyAndUser(ctx context.Context, rawKey string) (*domain.APIKey, *domain.User, error) {
	users, err := h.repos.User.FindAll(ctx)
	if err != nil {
		return nil, nil, err
	}

	for _, user := range users {
		keys, err := h.repos.APIKey.FindByUserID(ctx, user.ID)
		if err != nil {
			continue
		}
		for _, key := range keys {
			if err := bcrypt.CompareHashAndPassword([]byte(key.KeyHash), []byte(rawKey)); err == nil {
				return key, user, nil
			}
		}
	}
	return nil, nil, nil
}

// Session handles login via token (from CLI)
func (h *AuthHandler) Session(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, `{"error": "missing token"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Find web session by token
	webSession, err := h.repos.WebSession.FindByToken(ctx, token)
	if err != nil || webSession == nil {
		http.Error(w, `{"error": "invalid token"}`, http.StatusUnauthorized)
		return
	}

	// Check if session is expired
	if webSession.IsExpired() {
		_ = h.repos.WebSession.Delete(ctx, webSession.ID)
		http.Error(w, `{"error": "token expired"}`, http.StatusUnauthorized)
		return
	}

	// Create a new session with longer duration
	sessionToken, err := generateToken()
	if err != nil {
		http.Error(w, `{"error": "failed to generate session token"}`, http.StatusInternalServerError)
		return
	}

	newSession := &domain.WebSession{
		UserID:    webSession.UserID,
		Token:     sessionToken,
		ExpiresAt: time.Now().Add(sessionDuration),
	}
	if err := h.repos.WebSession.Create(ctx, newSession); err != nil {
		http.Error(w, `{"error": "failed to create web session"}`, http.StatusInternalServerError)
		return
	}

	// Delete the old token
	_ = h.repos.WebSession.Delete(ctx, webSession.ID)

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionToken,
		Path:     "/",
		Expires:  newSession.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to dashboard (use WEB_URL if set)
	redirectURL := "/"
	if h.cfg.WebURL != "" {
		redirectURL = h.cfg.WebURL
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// CreateWebSession creates a web session for CLI login
func (h *AuthHandler) CreateWebSession(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, `{"error": "user not found in context"}`, http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	// Generate token
	token, err := generateToken()
	if err != nil {
		http.Error(w, `{"error": "failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(webSessionDuration)

	webSession := &domain.WebSession{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	if err := h.repos.WebSession.Create(ctx, webSession); err != nil {
		http.Error(w, `{"error": "failed to create web session"}`, http.StatusInternalServerError)
		return
	}

	// Build URL
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s/auth/session?token=%s", scheme, r.Host, token)

	resp := WebSessionResponse{
		URL:       url,
		ExpiresAt: expiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok": true}`))
		return
	}

	ctx := r.Context()

	// Find and delete session
	session, err := h.repos.WebSession.FindByToken(ctx, cookie.Value)
	if err == nil && session != nil {
		_ = h.repos.WebSession.Delete(ctx, session.ID)
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok": true}`))
}

// Me returns the current user
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, `{"error": "user not found"}`, http.StatusUnauthorized)
		return
	}

	resp := MeResponse{User: user}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListUsers returns all users
func (h *AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := h.repos.User.FindAll(ctx)
	if err != nil {
		http.Error(w, `{"error": "failed to list users"}`, http.StatusInternalServerError)
		return
	}

	resp := UsersResponse{Users: users}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListKeys returns API keys for the current user
func (h *AuthHandler) ListKeys(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, `{"error": "user not found"}`, http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	keys, err := h.repos.APIKey.FindByUserID(ctx, user.ID)
	if err != nil {
		http.Error(w, `{"error": "failed to list keys"}`, http.StatusInternalServerError)
		return
	}

	keyInfos := make([]*APIKeyInfo, len(keys))
	for i, k := range keys {
		keyInfos[i] = &APIKeyInfo{
			ID:         k.ID,
			Name:       k.Name,
			KeyPrefix:  k.KeyPrefix,
			LastUsedAt: k.LastUsedAt,
			CreatedAt:  k.CreatedAt,
		}
	}

	resp := ListKeysResponse{Keys: keyInfos}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreateKey creates a new API key for the current user
func (h *AuthHandler) CreateKey(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, `{"error": "user not found"}`, http.StatusUnauthorized)
		return
	}

	var req CreateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"error": "name is required"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Generate API key
	rawKey, err := generateAPIKey()
	if err != nil {
		http.Error(w, `{"error": "failed to generate api key"}`, http.StatusInternalServerError)
		return
	}

	keyHash, err := hashAPIKey(rawKey)
	if err != nil {
		http.Error(w, `{"error": "failed to hash api key"}`, http.StatusInternalServerError)
		return
	}

	apiKey := &domain.APIKey{
		UserID:    user.ID,
		Name:      req.Name,
		KeyHash:   keyHash,
		KeyPrefix: rawKey[:12] + "...",
	}
	if err := h.repos.APIKey.Create(ctx, apiKey); err != nil {
		http.Error(w, `{"error": "failed to create api key"}`, http.StatusInternalServerError)
		return
	}

	resp := CreateKeyResponse{
		Key: &APIKeyInfo{
			ID:        apiKey.ID,
			Name:      apiKey.Name,
			KeyPrefix: apiKey.KeyPrefix,
			CreatedAt: apiKey.CreatedAt,
		},
		APIKey: rawKey,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeleteKey deletes an API key
func (h *AuthHandler) DeleteKey(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, `{"error": "user not found"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	keyID := vars["id"]
	if keyID == "" {
		http.Error(w, `{"error": "key id is required"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Find key and verify ownership
	key, err := h.repos.APIKey.FindByID(ctx, keyID)
	if err != nil || key == nil {
		http.Error(w, `{"error": "key not found"}`, http.StatusNotFound)
		return
	}

	if key.UserID != user.ID {
		http.Error(w, `{"error": "forbidden"}`, http.StatusForbidden)
		return
	}

	if err := h.repos.APIKey.Delete(ctx, keyID); err != nil {
		http.Error(w, `{"error": "failed to delete key"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok": true}`))
}
