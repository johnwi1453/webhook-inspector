package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"webhook-inspector/internal/redis"
)

func GetTokenStatus(w http.ResponseWriter, r *http.Request) {
	token, ok := GetToken(w, r)
	if !ok {
		return
	}

	countKey := fmt.Sprintf("hooks:%s:count", token)

	// Get usage count
	count, err := redis.Client.Get(context.Background(), countKey).Int()
	if err != nil && err.Error() != "redis: nil" {
		http.Error(w, "failed to fetch usage count", http.StatusInternalServerError)
		return
	}
	if err != nil {
		count = 0
	}

	// Get TTL
	ttl, err := redis.Client.TTL(context.Background(), countKey).Result()
	if err != nil {
		http.Error(w, "failed to fetch TTL", http.StatusInternalServerError)
		return
	}

	// Check if token has a privileged owner
	owner, err := redis.Client.Get(context.Background(), "token:"+token+":owner").Result()
	isPrivileged := (err == nil && owner != "")

	resp := map[string]interface{}{
		"token":              token,
		"requests_used":      count,
		"requests_remaining": max(0, 5-int(count)),
		"limit":              5,
		"ttl_seconds":        int(ttl.Seconds()),
		"owner":              owner,
		"privileged":         isPrivileged,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
