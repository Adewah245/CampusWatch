package model

import "time"

type Issue struct {
	ID          string    `json:"id"`
	SystemID    string    `json:"system_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	ReportedBy  string    `json:"reported_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateIssueRequest struct {
	SystemID    string `json:"system_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority,omitempty"`
	ReportedBy  string `json:"reported_by,omitempty"`
}

type UpdateIssueRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Priority    *string `json:"priority,omitempty"`
	Status      *string `json:"status,omitempty"`
}

type IssueUpdate struct {
	ID        string    `json:"id"`
	IssueID   string    `json:"issue_id"`
	Status    *string   `json:"status,omitempty"`
	Comment   string    `json:"comment"`
	UpdatedBy string    `json:"updated_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateIssueUpdateRequest struct {
	Status    *string `json:"status,omitempty"`
	Comment   string  `json:"comment"`
	UpdatedBy string  `json:"updated_by,omitempty"`
}
