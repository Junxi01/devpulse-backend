package mock

import (
	"encoding/json"
	"time"
)

// WebhookEnvelope mirrors a minimal GitHub webhook delivery wrapper used for local mock JSON files.
type WebhookEnvelope struct {
	DeliveryID string          `json:"delivery_id"`
	EventType  string          `json:"event_type"`
	OccurredAt time.Time       `json:"occurred_at"`
	Payload    json.RawMessage `json:"payload"`
}
