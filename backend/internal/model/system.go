package model

import "time"

// System represents a monitored computer assigned to a physical location.
type System struct {
	ID              string     `json:"id"`
	LocationID      string     `json:"location_id"`
	Hostname        string     `json:"hostname"`
	DeviceName      string     `json:"device_name"`
	OperatingSystem string     `json:"operating_system"`
	OSVersion       string     `json:"os_version,omitempty"`
	SystemStatus    string     `json:"system_status"`
	MonitorStatus   string     `json:"monitor_status"`
	LastSeenAt      *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateSystemRequest struct {
	LocationID      string `json:"location_id"`
	Hostname        string `json:"hostname"`
	DeviceName      string `json:"device_name"`
	OperatingSystem string `json:"operating_system"`
	OSVersion       string `json:"os_version,omitempty"`
}

type UpdateSystemRequest struct {
	LocationID      *string `json:"location_id,omitempty"`
	Hostname        *string `json:"hostname,omitempty"`
	DeviceName      *string `json:"device_name,omitempty"`
	OperatingSystem *string `json:"operating_system,omitempty"`
	OSVersion       *string `json:"os_version,omitempty"`
	SystemStatus    *string `json:"system_status,omitempty"`
	MonitorStatus   *string `json:"monitor_status,omitempty"`
}
