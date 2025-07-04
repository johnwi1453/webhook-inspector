package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"webhook-inspector/internal/config"
	"webhook-inspector/internal/redis"
)

func GetTokenStatus(w http.ResponseWriter, r *http.Request) {
	token, ok := GetToken(w, r)
	if !ok {
		return
	}

	countKey := fmt.Sprintf("rate_limit:%s", token)

	// Get usage count
	count, err := redis.Client.Get(context.Background(), countKey).Int()
	if err != nil && err.Error() != "redis: nil" {
		log.Printf("GetTokenStatus: failed to fetch usage count for token %s: %v", token, err)
		http.Error(w, "failed to fetch usage count", http.StatusInternalServerError)
		return
	}
	if err != nil {
		count = 0
	}

	// Get TTL
	ttl, err := redis.Client.TTL(context.Background(), countKey).Result()
	if err != nil {
		log.Printf("GetTokenStatus: failed to fetch TTL for token %s: %v", token, err)
		http.Error(w, "failed to fetch TTL", http.StatusInternalServerError)
		return
	}

	// Check if token has a privileged owner
	owner, err := redis.Client.Get(context.Background(), "token:"+token+":owner").Result()
	isPrivileged := false
	if err == nil && owner != "" {
		isPrivileged = true

		// Optional: verify that session_token matches owner
		if sessionCookie, err := r.Cookie("session_token"); err == nil {
			sessionToken := sessionCookie.Value
			username, _ := redis.Client.Get(context.Background(), "user:"+sessionToken).Result()
			if username != owner {
				isPrivileged = false // logged-in user mismatch
			}
		}
	}

	maxLimit := config.AnonymousRateLimit
	if isPrivileged {
		maxLimit = config.PrivilegedRateLimit
	}

	resp := map[string]interface{}{
		"token":              token,
		"requests_used":      count,
		"requests_remaining": max(0, maxLimit-int(count)),
		"limit":              maxLimit,
		"ttl":                fmt.Sprintf("%dh %dm", int(ttl.Hours()), int(ttl.Minutes())%60),
		"owner":              owner,
		"privileged":         isPrivileged,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
