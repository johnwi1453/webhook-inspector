package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestWebhookPayload_MarshalUnmarshal(t *testing.T) {
	original := WebhookPayload{
		ID:        "abc123",
		Method:    "POST",
		Headers:   map[string][]string{"Content-Type": {"application/json"}},
		Body:      `{"foo":"bar"}`,
		Timestamp: time.Now().UTC().Truncate(time.Second),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded WebhookPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.ID != original.ID ||
		decoded.Method != original.Method ||
		decoded.Body != original.Body ||
		decoded.Headers["Content-Type"][0] != "application/json" {
		t.Errorf("Decoded struct does not match original")
	}
}
