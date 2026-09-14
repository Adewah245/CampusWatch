package model

import (
	"encoding/json"
	"time"
)

type Event struct {
	ID         string          `json:"id"`
	SystemID   string          `json:"system_id"`
	AgentID    *string         `json:"agent_id,omitempty"`
	EventType  string          `json:"event_type"`
	Payload    json.RawMessage `json:"payload"`
	OccurredAt time.Time       `json:"occurred_at"`
	CreatedAt  time.Time       `json:"created_at"`
}

type EventRequest struct {
	AgentID    string          `json:"agent_id"`
	SystemID   string          `json:"system_id"`
	EventType  string          `json:"event_type"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	OccurredAt time.Time       `json:"occurred_at"`
}
