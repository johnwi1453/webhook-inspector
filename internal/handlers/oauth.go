package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"webhook-inspector/internal/auth"
	"webhook-inspector/internal/redis"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

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
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	// Fetch user info from GitHub
	client := auth.GithubOAuthConfig.Client(r.Context(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil || resp.StatusCode != 200 {
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

	// Generate new token (can be mapped to GitHub ID)
	sessionToken := uuid.New().String()
	redis.Client.Set(r.Context(), "user:"+sessionToken, ghUser.Login, 0)

	// Set secure cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
	})

	fmt.Fprintf(w, "✅ Logged in as %s!", ghUser.Login)
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
