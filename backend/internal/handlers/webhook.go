package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"webhook-inspector/internal/config"
	"webhook-inspector/internal/models"
	"webhook-inspector/internal/redis"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

// Helpers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Helper function to sanitize sensitive data from logs
func sanitizeForLogging(data string) string {
	// List of sensitive field patterns to redact
	sensitivePatterns := []string{
		"password", "token", "secret", "key", "auth", "credential",
		"bearer", "api_key", "apikey", "access_token", "refresh_token",
		"client_secret", "private_key", "ssh_key", "certificate",
	}

	// If data is too long, truncate it
	if len(data) > 500 {
		data = data[:500] + "...[truncated]"
	}

	// Check if any sensitive patterns are present (case insensitive)
	lowerData := strings.ToLower(data)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(lowerData, pattern) {
			return "[REDACTED - contains sensitive data]"
		}
	}

	return data
}

// Store webhook in Redis
func HandleWebhook(w http.ResponseWriter, r *http.Request) {
	token, ok := GetToken(w, r)
	if !ok {
		return
	}

	// Check privilege status and set rate limit
	owner, err := redis.Client.Get(context.Background(), "token:"+token+":owner").Result()
	isPrivileged := (err == nil && owner != "")

	maxRequestsPerToken := config.AnonymousRateLimit
	if isPrivileged {
		maxRequestsPerToken = config.PrivilegedRateLimit
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("HandleWebhook: failed to read request body for token %s: %v", token, err)
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	// Validate JSON
	var parsedBody interface{}
	if err := json.Unmarshal(bodyBytes, &parsedBody); err != nil {
		// Sanitize body data before logging
		sanitizedBody := sanitizeForLogging(string(bodyBytes))
		log.Printf("HandleWebhook: invalid JSON body for token %s: %v, body: %s", token, err, sanitizedBody)
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
		log.Printf("HandleWebhook: failed to marshal payload for token %s: %v", token, err)
		http.Error(w, "failed parse request body into json", http.StatusInternalServerError)
		return
	}

	countKey := fmt.Sprintf("rate_limit:%s", token)

	// Use Redis pipeline for atomic increment and TTL setting
	pipe := redis.Client.Pipeline()
	incrCmd := pipe.Incr(context.Background(), countKey)

	// Always set TTL to ensure it doesn't get lost
	pipe.Expire(context.Background(), countKey, config.RateLimitTTL)

	_, err = pipe.Exec(context.Background())
	if err != nil {
		log.Printf("HandleWebhook: failed to execute rate limit pipeline for token %s: %v", token, err)
		http.Error(w, "failed to track webhook usage", http.StatusInternalServerError)
		return
	}

	count := int(incrCmd.Val())

	if count > maxRequestsPerToken {
		log.Printf("Token %s blocked (rate limit %d)", token, count)
		http.Error(w, "rate limit exceeded for this token", http.StatusTooManyRequests)
		return
	}

	// Write webhook into redis
	err = redis.Client.Set(context.Background(), key, jsonData, config.WebhookDataTTL).Err()
	if err != nil {
		log.Printf("HandleWebhook: failed to save webhook for token %s: %v", token, err)
		http.Error(w, "failed to save webhook", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Saved webhook with ID %s for token %s\n", id, token)
	remaining := max(0, maxRequestsPerToken-int(count))
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
		log.Printf("GetWebhookLogs: failed to fetch keys for token %s: %v", token, err)
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
	// First, check if user is logged in via session_token
	if sessionCookie, err := r.Cookie("session_token"); err == nil {
		sessionToken := sessionCookie.Value
		username, err := redis.Client.Get(context.Background(), "user:"+sessionToken).Result()
		if err == nil && username != "" {
			// GitHub user: use or create privileged token
			existingToken, err := redis.Client.Get(context.Background(), "user:"+username+":webhook_token").Result()
			if err == goredis.Nil {
				existingToken = uuid.New().String()
				redis.Client.Set(context.Background(), "user:"+username+":webhook_token", existingToken, 0)
				redis.Client.Set(context.Background(), "token:"+existingToken+":owner", username, 0)
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "webhook_token",
				Value:    existingToken,
				Path:     "/",
				HttpOnly: true,
				Secure:   isSecureContext(r),
				SameSite: http.SameSiteLaxMode,
				MaxAge:   config.SessionCookieTTL,
			})

			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(fmt.Sprintf("Assigned privileged token: %s", existingToken)))
			return
		}
	}

	// If not logged in: generate random anonymous token
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

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf("Assigned new anonymous token: %s", newToken)))
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

// Delete individual webhooks
func DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	token, ok := GetToken(w, r)
	if !ok {
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing webhook ID", http.StatusBadRequest)
		return
	}

	key := fmt.Sprintf("hooks:%s:%s", token, id)

	err := redis.Client.Del(context.Background(), key).Err()
	if err != nil {
		log.Printf("DeleteWebhook: failed to delete webhook %s for token %s: %v", id, token, err)
		http.Error(w, "Failed to delete webhook", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Deleted"))
}
