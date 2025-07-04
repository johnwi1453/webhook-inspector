package config

import (
	"os"
	"strconv"
	"time"
)

// Rate limiting configuration
var (
	// Rate limits
	AnonymousRateLimit  = getEnvInt("ANONYMOUS_RATE_LIMIT", 50)
	PrivilegedRateLimit = getEnvInt("PRIVILEGED_RATE_LIMIT", 500)

	// TTL durations
	WebhookDataTTL   = getEnvDuration("WEBHOOK_DATA_TTL", 24*time.Hour)
	RateLimitTTL     = getEnvDuration("RATE_LIMIT_TTL", 24*time.Hour)
	SessionCookieTTL = getEnvInt("SESSION_COOKIE_TTL", 86400*3) // 3 days in seconds
)

// Helper function to get environment variable as int with default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Helper function to get environment variable as duration with default
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
