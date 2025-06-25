package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"time"

	"webhook-inspector/internal/models"
	"webhook-inspector/internal/redis"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Helpers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

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

	owner, err := redis.Client.Get(context.Background(), "token:"+token+":owner").Result()
	isPrivileged := (err == nil && owner != "")

	MaxRequestsPerToken := 5
	if isPrivileged {
		MaxRequestsPerToken = 500
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	// Validate JSON
	var parsedBody interface{}
	if err := json.Unmarshal(bodyBytes, &parsedBody); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Generate uuid and key for redis storage
	id := uuid.New().String()
	key := fmt.Sprintf("hooks:%s:%s", token, id)

	// Map request to webhook data model
	payload := models.WebhookPayload{
		ID:        id,
		Method:    r.Method,
		Headers:   r.Header,
		Body:      string(bodyBytes),
		Timestamp: time.Now().UTC(),
	}

	// Format into json data
	jsonData, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "failed parse request body into json", http.StatusInternalServerError)
		return
	}

	countKey := fmt.Sprintf("hooks:%s:count", token)

	count64, err := redis.Client.Incr(context.Background(), countKey).Result()
	if err != nil {
		http.Error(w, "failed to track webhook usage", http.StatusInternalServerError)
		return
	}
	count := int(count64)

	if count == 1 {
		// first time we've seen this token — set TTL for 24h
		redis.Client.Expire(context.Background(), countKey, 24*time.Hour)
	}

	if count > MaxRequestsPerToken {
		log.Printf("Token %s blocked (rate limit %d)", token, count)
		http.Error(w, "rate limit exceeded for this token", http.StatusTooManyRequests)
		return
	}

	// Write webhook into redis
	err = redis.Client.Set(context.Background(), key, jsonData, 24*time.Hour).Err()
	if err != nil {
		http.Error(w, "failed to save webhook", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Saved webhook with ID %s for token %s\n", id, token)
	remaining := max(0, MaxRequestsPerToken-int(count))
	w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
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

	var logs []models.WebhookPayload

	// Get logs matching our pattern into logs
	for _, key := range keys {
		val, err := redis.Client.Get(context.Background(), key).Result()
		if err != nil {
			continue
		}

		var parsed models.WebhookPayload
		if err := json.Unmarshal([]byte(val), &parsed); err != nil {
			continue // skip invalid entries
		}

		logs = append(logs, parsed)
	}

	// Sort by timestamp (oldest first)
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.Before(logs[j].Timestamp)
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(logs)

}

// Create new session and token for new users
func CreateSession(w http.ResponseWriter, r *http.Request) {
	// Check for an existing token in browser cookies
	oldCookie, err := r.Cookie("webhook_token")
	if err == nil {
		oldToken := oldCookie.Value
		pattern := fmt.Sprintf("hooks:%s:*", oldToken)

		keys, err := redis.Client.Keys(context.Background(), pattern).Result()
		if err == nil && len(keys) > 0 {
			if err := redis.Client.Del(context.Background(), keys...).Err(); err == nil {
				log.Printf("🔄 Deleted old token and %d keys for token %s\n", len(keys), oldToken)
			} else {
				log.Printf("Failed to delete keys for old token %s: %v\n", oldToken, err)
			}
		}
	}

	// Generate and assign new token
	newToken := uuid.New().String()

	http.SetCookie(w, &http.Cookie{
		Name:     "webhook_token",
		Value:    newToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 3, // 3 days
	})

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf(`
	Your webhook token has been created!

	Use these endpoints:
	- POST to /api/hooks/%s
	- GET from /logs/%s
	`, newToken, newToken)))
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
