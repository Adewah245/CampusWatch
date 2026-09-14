package model

type DashboardSummary struct {
	TotalSystems       int `json:"total_systems"`
	OnlineSystems      int `json:"online_systems"`
	OfflineSystems     int `json:"offline_systems"`
	InactiveSystems    int `json:"inactive_systems"`
	MaintenanceSystems int `json:"maintenance_systems"`
	ActiveUsers        int `json:"active_users"`
	OpenIssues         int `json:"open_issues"`
	AfterHoursUsage    int `json:"after_hours_usage"`
}
