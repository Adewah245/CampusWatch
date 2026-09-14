package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"CampusWatch/backend/internal/model"
)

type IssueRepository interface {
	Create(context.Context, model.Issue) (*model.Issue, error)
	List(context.Context) ([]model.Issue, error)
	GetByID(context.Context, string) (*model.Issue, error)
	Update(context.Context, string, model.UpdateIssueRequest) (*model.Issue, error)
	CreateUpdate(context.Context, string, model.CreateIssueUpdateRequest) (*model.IssueUpdate, error)
}

type IssueService struct{ repo IssueRepository }

func NewIssueService(repo IssueRepository) *IssueService { return &IssueService{repo: repo} }

func (s *IssueService) Create(ctx context.Context, input model.CreateIssueRequest) (*model.Issue, error) {
	input.SystemID, input.Title, input.Description = strings.TrimSpace(input.SystemID), strings.TrimSpace(input.Title), strings.TrimSpace(input.Description)
	if input.SystemID == "" {
		return nil, errors.New("system id is required")
	}
	if input.Title == "" {
		return nil, errors.New("issue title is required")
	}
	if input.Description == "" {
		return nil, errors.New("issue description is required")
	}
	if input.Priority == "" {
		input.Priority = "medium"
	}
	if !validIssuePriority(input.Priority) {
		return nil, fmt.Errorf("invalid issue priority %q", input.Priority)
	}
	return s.repo.Create(ctx, model.Issue{SystemID: input.SystemID, Title: input.Title, Description: input.Description, Priority: input.Priority, Status: "open", ReportedBy: strings.TrimSpace(input.ReportedBy)})
}

func (s *IssueService) List(ctx context.Context) ([]model.Issue, error) { return s.repo.List(ctx) }
func (s *IssueService) GetByID(ctx context.Context, id string) (*model.Issue, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("issue id is required")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *IssueService) Update(ctx context.Context, id string, input model.UpdateIssueRequest) (*model.Issue, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("issue id is required")
	}
	if input.Title != nil {
		value := strings.TrimSpace(*input.Title)
		if value == "" {
			return nil, errors.New("issue title cannot be empty")
		}
		input.Title = &value
	}
	if input.Description != nil {
		value := strings.TrimSpace(*input.Description)
		if value == "" {
			return nil, errors.New("issue description cannot be empty")
		}
		input.Description = &value
	}
	if input.Priority != nil {
		value := strings.ToLower(strings.TrimSpace(*input.Priority))
		if !validIssuePriority(value) {
			return nil, fmt.Errorf("invalid issue priority %q", value)
		}
		input.Priority = &value
	}
	if input.Status != nil {
		value := strings.ToLower(strings.TrimSpace(*input.Status))
		if !validIssueStatus(value) {
			return nil, fmt.Errorf("invalid issue status %q", value)
		}
		input.Status = &value
	}
	return s.repo.Update(ctx, id, input)
}

func (s *IssueService) AddUpdate(ctx context.Context, id string, input model.CreateIssueUpdateRequest) (*model.IssueUpdate, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("issue id is required")
	}
	input.Comment = strings.TrimSpace(input.Comment)
	if input.Comment == "" {
		return nil, errors.New("issue update comment is required")
	}
	if input.Status != nil {
		value := strings.ToLower(strings.TrimSpace(*input.Status))
		if !validIssueStatus(value) {
			return nil, fmt.Errorf("invalid issue status %q", value)
		}
		input.Status = &value
	}
	return s.repo.CreateUpdate(ctx, id, input)
}

func validIssuePriority(value string) bool {
	switch strings.ToLower(value) {
	case "low", "medium", "high", "critical":
		return true
	default:
		return false
	}
}
func validIssueStatus(value string) bool {
	switch strings.ToLower(value) {
	case "open", "acknowledged", "in_progress", "resolved", "closed":
		return true
	default:
		return false
	}
}
