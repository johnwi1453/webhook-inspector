package handlers

import (
	"context"
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

	// Delete webhook data
	pattern := fmt.Sprintf("hooks:%s:*", token)
	keys, _ := redis.Client.Keys(context.Background(), pattern).Result()
	if len(keys) > 0 {
		redis.Client.Del(context.Background(), keys...)
	}
	redis.Client.Del(context.Background(), fmt.Sprintf("hooks:%s:count", token))

	// Check if user is logged in via session_token
	var newToken string
	if sessionCookie, err := r.Cookie("session_token"); err == nil {
		sessionToken := sessionCookie.Value
		username, err := redis.Client.Get(context.Background(), "user:"+sessionToken).Result()
		if err == nil && username != "" {
			// GitHub user found — assign privileged token
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
		MaxAge:   86400 * 3,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true, "message": "Token reset complete"}`))
}
