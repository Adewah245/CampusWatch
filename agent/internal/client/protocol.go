// Package client speaks the CampusWatch agent API contract.
//
// The request and response types in this file deliberately mirror the JSON
// shapes the backend already decodes (see backend/internal/model/agent.go,
// health.go, user_session.go and event.go). They are declared here rather than
// imported from the backend module on purpose:
//
//   - README section 40 states the teams integrate through API contracts, not
//     through each other's implementation. Restating the contract keeps the
//     agent honest about what it depends on.
//   - The agent ships to monitored computers. Importing CampusWatch/backend
//     would drag the entire server source tree into every agent build.
//
// If the backend contract changes, these types must be updated in step.
package client

import (
	"encoding/json"
	"time"
)

// HealthReport carries the operational metrics collected from the monitored
// computer. It mirrors model.HealthReport on the backend.
//
// BatteryPercent is a pointer because desktops have no battery at all, which
// is meaningfully different from a battery reading of zero percent.
type HealthReport struct {
	SystemID         string    `json:"system_id,omitempty"`
	RecordedAt       time.Time `json:"recorded_at"`
	CPUPercent       float64   `json:"cpu_percent"`
	MemoryPercent    float64   `json:"memory_percent"`
	DiskPercent      float64   `json:"disk_percent"`
	BatteryPercent   *float64  `json:"battery_percent,omitempty"`
	NetworkConnected bool      `json:"network_connected"`
	OperatingSystem  string    `json:"operating_system,omitempty"`
	AgentHealth      string    `json:"agent_health"`
	UptimeSeconds    int64     `json:"uptime_seconds"`
}

// HeartbeatRequest is the body posted to POST /api/v1/agents/heartbeat.
type HeartbeatRequest struct {
	AgentID      string        `json:"agent_id"`
	SystemID     string        `json:"system_id"`
	Timestamp    time.Time     `json:"timestamp"`
	AgentVersion string        `json:"agent_version,omitempty"`
	Health       *HealthReport `json:"health,omitempty"`
}

// HeartbeatResponse is the acknowledgement returned by the heartbeat endpoint.
type HeartbeatResponse struct {
	AgentID    string    `json:"agent_id"`
	SystemID   string    `json:"system_id"`
	Status     string    `json:"status"`
	ReceivedAt time.Time `json:"received_at"`
}

// SessionEventRequest is the body posted to POST /api/v1/agents/session.
//
// Username is the operating system account, not the CampusWatch identity. The
// backend maps the OS account onto an institutional identity, which is how
// README section 7 resolves shared lab accounts such as "student".
type SessionEventRequest struct {
	AgentID   string    `json:"agent_id"`
	SystemID  string    `json:"system_id"`
	Username  string    `json:"username"`
	Event     string    `json:"event"`
	SessionID string    `json:"session_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// EventRequest is the body posted to POST /api/v1/agents/event for anything
// that is not a heartbeat or a user session transition, such as health
// warnings and fault detection.
type EventRequest struct {
	AgentID    string          `json:"agent_id"`
	SystemID   string          `json:"system_id"`
	EventType  string          `json:"event_type"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	OccurredAt time.Time       `json:"occurred_at"`
}

// User session event names accepted by POST /api/v1/agents/session.
const (
	EventUserLogin  = "USER_LOGIN"
	EventUserLogout = "USER_LOGOUT"
	EventUserIdle   = "USER_IDLE"
	EventUserActive = "USER_ACTIVE"
)

// Event names accepted by POST /api/v1/agents/event.
const (
	EventHealthWarning = "HEALTH_WARNING"
	EventFaultDetected = "FAULT_DETECTED"
)
