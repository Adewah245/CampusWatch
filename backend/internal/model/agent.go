package model

import "time"

type Agent struct {
	ID              string     `json:"id"`
	SystemID        string     `json:"system_id"`
	AgentCode       string     `json:"agent_code"`
	AgentVersion    string     `json:"agent_version"`
	Status          string     `json:"status"`
	LastHeartbeatAt *time.Time `json:"last_heartbeat_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type AgentRegistrationRequest struct {
	SystemID     string `json:"system_id"`
	AgentCode    string `json:"agent_code"`
	AgentVersion string `json:"agent_version"`
}

type AgentCredentialResponse struct {
	Agent      Agent  `json:"agent"`
	Credential string `json:"credential"`
}

type AgentAuth struct {
	Agent          Agent  `json:"-"`
	CredentialHash string `json:"-"`
}

type HeartbeatRequest struct {
	AgentID      string        `json:"agent_id"`
	SystemID     string        `json:"system_id"`
	Timestamp    time.Time     `json:"timestamp"`
	AgentVersion string        `json:"agent_version,omitempty"`
	Health       *HealthReport `json:"health,omitempty"`
}

type HeartbeatResponse struct {
	AgentID    string    `json:"agent_id"`
	SystemID   string    `json:"system_id"`
	Status     string    `json:"status"`
	ReceivedAt time.Time `json:"received_at"`
}
