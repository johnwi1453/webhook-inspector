package models

import (
	"time"
)

type WebhookPayload struct {
	ID        string              `json:"id"`
	Method    string              `json:"method"`
	Headers   map[string][]string `json:"headers"`
	Body      string              `json:"body"`
	Timestamp time.Time           `json:"timestamp"`
}
