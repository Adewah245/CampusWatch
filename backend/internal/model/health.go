package model

import "time"

type HealthReport struct {
	SystemID         string    `json:"system_id"`
	RecordedAt       time.Time `json:"recorded_at"`
	CPUPercent       float64   `json:"cpu_percent"`
	MemoryPercent    float64   `json:"memory_percent"`
	DiskPercent      float64   `json:"disk_percent"`
	BatteryPercent   *float64  `json:"battery_percent,omitempty"`
	NetworkConnected bool      `json:"network_connected"`
	OperatingSystem  string    `json:"operating_system,omitempty"`
	AgentHealth      string    `json:"agent_health"`
	UptimeSeconds    int64     `json:"uptime_seconds"`
	CreatedAt        time.Time `json:"created_at,omitempty"`
}
