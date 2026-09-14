package model

import "time"

type UserSession struct {
	ID              string     `json:"id"`
	SystemID        string     `json:"system_id"`
	Username        string     `json:"username"`
	State           string     `json:"state"`
	LoginAt         time.Time  `json:"login_at"`
	LogoutAt        *time.Time `json:"logout_at,omitempty"`
	LastActivityAt  time.Time  `json:"last_activity_at"`
	DurationSeconds *int64     `json:"duration_seconds,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type UserSessionEventRequest struct {
	AgentID   string    `json:"agent_id"`
	SystemID  string    `json:"system_id"`
	Username  string    `json:"username"`
	Event     string    `json:"event"`
	SessionID string    `json:"session_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
