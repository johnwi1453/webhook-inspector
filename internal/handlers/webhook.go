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

// Store webhook in Redis
func HandleWebhook(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

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
	token := chi.URLParam(r, "token")
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
