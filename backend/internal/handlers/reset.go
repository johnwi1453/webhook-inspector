package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"webhook-inspector/internal/redis"

	"github.com/google/uuid"
)

func ResetToken(w http.ResponseWriter, r *http.Request) {
	token, ok := GetToken(w, r)
	if !ok {
		return
	}

	// Delete all logs and rate limit counters
	pattern := fmt.Sprintf("hooks:%s:*", token)
	keys, _ := redis.Client.Keys(context.Background(), pattern).Result()
	if len(keys) > 0 {
		redis.Client.Del(context.Background(), keys...)
	}
	redis.Client.Del(context.Background(), fmt.Sprintf("hooks:%s:count", token))
	redis.Client.Del(context.Background(), fmt.Sprintf("token:%s:owner", token))

	// Generate new token
	newToken := uuid.New().String()

	// Check for GitHub session
	sessionCookie, err := r.Cookie("session_token")
	if err == nil {
		username, err := redis.Client.Get(context.Background(), "user:"+sessionCookie.Value).Result()
		if err == nil {
			// Replace GitHub user's token
			redis.Client.Set(context.Background(), "user:"+username+":webhook_token", newToken, 0)
			redis.Client.Set(context.Background(), "token:"+newToken+":owner", username, 0)
		}
	}

	// Overwrite the webhook_token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "webhook_token",
		Value:    newToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 3,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"new_token": newToken,
	})
}
