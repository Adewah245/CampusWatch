package model

import "time"

type Report struct {
	Kind             string    `json:"kind"`
	Period           string    `json:"period"`
	From             time.Time `json:"from"`
	To               time.Time `json:"to"`
	UserSessions     int       `json:"user_sessions"`
	UsageSeconds     int64     `json:"usage_seconds"`
	IssuesCreated    int       `json:"issues_created"`
	AfterHoursEvents int       `json:"after_hours_events"`
	UptimeSeconds    int64     `json:"uptime_seconds"`
	DowntimeSeconds  int64     `json:"downtime_seconds"`
}
