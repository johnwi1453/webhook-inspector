package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"webhook-inspector/internal/redis"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Home page with instructions
func Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
	Welcome to Webhook Inspector

	Use this tool to test and debug webhooks.
	- Visit /create to generate your own personal token.
	- Send POST requests to /api/hooks/{your_token}
	- View your received webhooks at /logs/{your_token}

	Each user has their own token saved in a cookie.
	`))
}

// Store webhook in Redis
func HandleWebhook(w http.ResponseWriter, r *http.Request) {
	token, ok := GetToken(w, r)
	if !ok {
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	id := uuid.New().String()

	key := fmt.Sprintf("hooks:%s:%s", token, id)

	err = redis.Client.Set(context.Background(), key, body, 24*time.Hour).Err()
	if err != nil {
		http.Error(w, "failed to save webhook", http.StatusInternalServerError)
		return
	}

	fmt.Printf("✅ Saved webhook with ID %s for token %s\n", id, token)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook received"))
}

// Get webhook from Redis
func GetWebhookLogs(w http.ResponseWriter, r *http.Request) {
	token, ok := GetToken(w, r)
	if !ok {
		return
	}

	pattern := fmt.Sprintf("hooks:%s:*", token)

	keys, err := redis.Client.Keys(context.Background(), pattern).Result()
	if err != nil {
		http.Error(w, "failed to fetch keys", http.StatusInternalServerError)
		return
	}

	var logs []map[string]interface{}
	for _, key := range keys {
		val, err := redis.Client.Get(context.Background(), key).Result()
		if err != nil {
			continue
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(val), &parsed); err != nil {
			continue // skip if the JSON can't be parsed
		}

		logs = append(logs, parsed)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(logs)
}

// Create new session and token for new users
func CreateSession(w http.ResponseWriter, r *http.Request) {
	token := uuid.New().String()

	http.SetCookie(w, &http.Cookie{
		Name:     "webhook_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 3, // 3 days
	})

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf(`
	✅ Your webhook token has been created!

	Use these endpoints:
	- POST to /api/hooks/%s
	- GET from /logs/%s
	`, token, token)))
}

// Force the user to use their assigned token
func GetToken(w http.ResponseWriter, r *http.Request) (string, bool) {
	urlToken := chi.URLParam(r, "token")
	cookie, err := r.Cookie("webhook_token")
	if err != nil {
		http.Error(w, "Missing webhook_token cookie", http.StatusForbidden)
		return "", false
	}

	if urlToken != "" && urlToken != cookie.Value {
		fmt.Printf("urlToken: %s cookie.Value: %s\n", urlToken, cookie.Value)
		http.Error(w, "Token mismatch", http.StatusForbidden)
		return "", false
	}

	return cookie.Value, true
}
