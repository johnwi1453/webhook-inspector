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

	// Delete all webhook data for this token
	pattern := fmt.Sprintf("hooks:%s:*", token)
	keys, _ := redis.Client.Keys(context.Background(), pattern).Result()
	if len(keys) > 0 {
		redis.Client.Del(context.Background(), keys...)
	}

	// Delete the count key
	redis.Client.Del(context.Background(), fmt.Sprintf("hooks:%s:count", token))

	// Generate a new token
	newToken := uuid.New().String()
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
