package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
	"webhook-inspector/internal/auth"
	"webhook-inspector/internal/config"
	"webhook-inspector/internal/redis"

	goredis "github.com/redis/go-redis/v9"

	"os"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// Helper function to determine if we should use Secure flag on cookies
func isSecureContext(r *http.Request) bool {
	// Use Secure flag if:
	// 1. Request is HTTPS, or
	// 2. We're in production (FRONTEND_URL contains https), or
	// 3. Explicitly set via environment variable
	if r.TLS != nil {
		return true
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL != "" && len(frontendURL) >= 5 && frontendURL[:5] == "https" {
		return true
	}

	// For development, allow insecure cookies
	return false
}

// Redirects user to GitHub login
func GitHubLogin(w http.ResponseWriter, r *http.Request) {
	url := auth.GithubOAuthConfig.AuthCodeURL("state-random", oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Handles callback from GitHub
func GitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code in callback", http.StatusBadRequest)
		return
	}

	token, err := auth.ExchangeCodeForToken(code)
	if err != nil {
		log.Printf("GitHubCallback: failed to exchange code for token: %v", err)
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	// Step 1: Get user info from GitHub
	client := auth.GithubOAuthConfig.Client(r.Context(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil || resp.StatusCode != 200 {
		log.Printf("GitHubCallback: failed to get user info from GitHub API: %v, status: %d", err, resp.StatusCode)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var ghUser struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
		Email string `json:"email"`
	}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &ghUser)

	// Step 2: Look up or create a webhook token for this user
	var finalToken string

	existingToken, err := redis.Client.Get(r.Context(), "user:"+ghUser.Login+":webhook_token").Result()
	if err == goredis.Nil {
		// If no token, generate a new one
		finalToken = uuid.New().String()
		redis.Client.Set(r.Context(), "user:"+ghUser.Login+":webhook_token", finalToken, 0)
		redis.Client.Set(r.Context(), "token:"+finalToken+":owner", ghUser.Login, 0)
	} else if err == nil {
		// If token exists, ensure owner is registered
		finalToken = existingToken
		redis.Client.Set(r.Context(), "token:"+finalToken+":owner", ghUser.Login, 0)
	} else {
		log.Printf("GitHubCallback: Redis error getting webhook token for user %s: %v", ghUser.Login, err)
		http.Error(w, "Redis error", http.StatusInternalServerError)
		return
	}

	// Step 3: Set webhook_token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "webhook_token",
		Value:    finalToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureContext(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   config.SessionCookieTTL,
	})

	// Step 4: Set session_token cookie with proper expiration
	sessionToken := uuid.New().String()
	sessionTTL := time.Duration(config.SessionCookieTTL) * time.Second
	redis.Client.Set(r.Context(), "user:"+sessionToken, ghUser.Login, sessionTTL)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureContext(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   config.SessionCookieTTL,
	})

	// Step 5: Redirect to frontend
	redirect := os.Getenv("FRONTEND_URL")
	if redirect == "" {
		redirect = "http://localhost:5173/dashboard"
	}
	http.Redirect(w, r, redirect+"?login=1", http.StatusFound)
}

// Get the info of the current logged-in user
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "No session token", http.StatusUnauthorized)
		return
	}

	username, err := redis.Client.Get(r.Context(), "user:"+cookie.Value).Result()
	if err != nil {
		http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
		return
	}

	resp := map[string]interface{}{
		"logged_in": true,
		"username":  username,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Logout the user and delete their token
func Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1, // Expire immediately
		Expires:  time.Unix(0, 0),
	})

	newToken := uuid.New().String()

	http.SetCookie(w, &http.Cookie{
		Name:     "webhook_token",
		Value:    newToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureContext(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   config.SessionCookieTTL,
	})

	redirect := os.Getenv("FRONTEND_URL")
	if redirect == "" {
		redirect = "http://localhost:5173/dashboard"
	}

	http.Redirect(w, r, redirect+"?logout=1", http.StatusFound)
}
