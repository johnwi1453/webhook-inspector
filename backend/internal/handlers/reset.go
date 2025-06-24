package handlers

import (
	"context"
	"fmt"
	"net/http"

	"webhook-inspector/internal/redis"
)

func ResetToken(w http.ResponseWriter, r *http.Request) {
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

	if len(keys) > 0 {
		if err := redis.Client.Del(context.Background(), keys...).Err(); err != nil {
			http.Error(w, "failed to delete keys", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook data and usage reset"))
}
