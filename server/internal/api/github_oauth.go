package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

const (
	githubAuthorizeURL = "https://github.com/login/oauth/authorize"
	githubTokenURL     = "https://github.com/login/oauth/access_token"
	githubUserURL      = "https://api.github.com/user"
)

// GitHubUser represents the user info from GitHub API
type GitHubUser struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// GitHubAuth initiates the GitHub OAuth flow
func (h *AuthHandler) GitHubAuth(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.IsGitHubOAuthEnabled() {
		http.Error(w, `{"error": "GitHub OAuth is not configured"}`, http.StatusNotImplemented)
		return
	}

	// Build GitHub authorize URL
	params := url.Values{}
	params.Set("client_id", h.cfg.GitHubClientID)
	params.Set("scope", "user:email")

	// Get the return URL from query params if provided
	returnTo := r.URL.Query().Get("returnTo")
	if returnTo != "" {
		params.Set("state", returnTo)
	}

	redirectURL := fmt.Sprintf("%s?%s", githubAuthorizeURL, params.Encode())
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// GitHubCallback handles the callback from GitHub OAuth
func (h *AuthHandler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.IsGitHubOAuthEnabled() {
		http.Error(w, `{"error": "GitHub OAuth is not configured"}`, http.StatusNotImplemented)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, `{"error": "missing code parameter"}`, http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state") // returnTo URL if provided

	// Exchange code for access token
	accessToken, err := h.exchangeCodeForToken(code)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to exchange code: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// Get user info from GitHub
	githubUser, err := h.getGitHubUser(accessToken)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to get user info: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	providerID := fmt.Sprintf("%d", githubUser.ID)

	// Check if OAuth connection already exists
	conn, err := h.repos.OAuthConnection.FindByProviderAndProviderID(ctx, "github", providerID)
	if err != nil {
		http.Error(w, `{"error": "database error"}`, http.StatusInternalServerError)
		return
	}

	var user *domain.User

	if conn != nil {
		// Existing user - get user info
		user, err = h.repos.User.FindByID(ctx, conn.UserID)
		if err != nil || user == nil {
			http.Error(w, `{"error": "user not found"}`, http.StatusInternalServerError)
			return
		}
	} else {
		// New user - create account
		email := githubUser.Email
		if email == "" {
			email = fmt.Sprintf("%s@github.local", githubUser.Login)
		}

		// Check if email already exists
		existingUser, err := h.repos.User.FindByEmail(ctx, email)
		if err != nil {
			http.Error(w, `{"error": "database error"}`, http.StatusInternalServerError)
			return
		}

		if existingUser != nil {
			// Link to existing user
			user = existingUser
		} else {
			// Create new user
			displayName := githubUser.Name
			if displayName == "" {
				displayName = githubUser.Login
			}

			user = &domain.User{
				ID:          uuid.New().String(),
				Email:       email,
				DisplayName: displayName,
				CreatedAt:   time.Now(),
			}

			if err := h.repos.User.Create(ctx, user); err != nil {
				http.Error(w, `{"error": "failed to create user"}`, http.StatusInternalServerError)
				return
			}
		}

		// Create OAuth connection
		oauthConn := &domain.OAuthConnection{
			ID:         uuid.New().String(),
			UserID:     user.ID,
			Provider:   "github",
			ProviderID: providerID,
			CreatedAt:  time.Now(),
		}

		if err := h.repos.OAuthConnection.Create(ctx, oauthConn); err != nil {
			http.Error(w, `{"error": "failed to create oauth connection"}`, http.StatusInternalServerError)
			return
		}
	}

	// Create web session
	if err := h.createSessionAndRedirect(ctx, w, r, user, state); err != nil {
		http.Error(w, `{"error": "failed to create session"}`, http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandler) exchangeCodeForToken(code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", h.cfg.GitHubClientID)
	data.Set("client_secret", h.cfg.GitHubClientSecret)
	data.Set("code", code)

	req, err := http.NewRequest("POST", githubTokenURL, nil)
	if err != nil {
		return "", err
	}
	req.URL.RawQuery = data.Encode()
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if result.Error != "" {
		return "", fmt.Errorf("github error: %s", result.Error)
	}

	return result.AccessToken, nil
}

func (h *AuthHandler) getGitHubUser(accessToken string) (*GitHubUser, error) {
	req, err := http.NewRequest("GET", githubUserURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (h *AuthHandler) createSessionAndRedirect(ctx context.Context, w http.ResponseWriter, r *http.Request, user *domain.User, returnTo string) error {
	// Generate session token
	sessionToken, err := generateToken()
	if err != nil {
		return err
	}

	// Create web session
	session := &domain.WebSession{
		UserID:    user.ID,
		Token:     sessionToken,
		ExpiresAt: time.Now().Add(sessionDuration),
	}
	if err := h.repos.WebSession.Create(ctx, session); err != nil {
		return err
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionToken,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// Determine redirect URL
	redirectURL := "/"
	if h.cfg.WebURL != "" {
		redirectURL = h.cfg.WebURL
	}
	if returnTo != "" && returnTo[0] == '/' {
		if h.cfg.WebURL != "" {
			redirectURL = h.cfg.WebURL + returnTo
		} else {
			redirectURL = returnTo
		}
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
	return nil
}
