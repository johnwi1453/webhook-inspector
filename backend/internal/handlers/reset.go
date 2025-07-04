package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"webhook-inspector/internal/config"
	"webhook-inspector/internal/redis"

	"github.com/google/uuid"
)

func ResetToken(w http.ResponseWriter, r *http.Request) {
	token, ok := GetToken(w, r)
	if !ok {
		return
	}

	// Delete webhook data
	pattern := fmt.Sprintf("hooks:%s:*", token)
	keys, err := redis.Client.Keys(context.Background(), pattern).Result()
	if err != nil {
		log.Printf("ResetToken: failed to get keys for pattern %s: %v", pattern, err)
		http.Error(w, "Failed to reset token", http.StatusInternalServerError)
		return
	}
	if len(keys) > 0 {
		err = redis.Client.Del(context.Background(), keys...).Err()
		if err != nil {
			log.Printf("ResetToken: failed to delete keys for token %s: %v", token, err)
			http.Error(w, "Failed to reset token", http.StatusInternalServerError)
			return
		}
	}
	err = redis.Client.Del(context.Background(), fmt.Sprintf("rate_limit:%s", token)).Err()
	if err != nil {
		log.Printf("ResetToken: failed to delete rate limit key for token %s: %v", token, err)
	}

	// Check if user is logged in via session_token
	var newToken string
	if sessionCookie, err := r.Cookie("session_token"); err == nil {
		sessionToken := sessionCookie.Value
		username, err := redis.Client.Get(context.Background(), "user:"+sessionToken).Result()
		if err == nil && username != "" {
			// GitHub user found, so assign user-associated token
			newToken = uuid.New().String()
			redis.Client.Set(context.Background(), "user:"+username+":webhook_token", newToken, 0)
			redis.Client.Set(context.Background(), "token:"+newToken+":owner", username, 0)
		}
	}

	// If not logged in or failed, fallback to anonymous token
	if newToken == "" {
		newToken = uuid.New().String()
	}

	// Set new token in cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "webhook_token",
		Value:    newToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   config.SessionCookieTTL,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true, "message": "Token reset complete"}`))
}
