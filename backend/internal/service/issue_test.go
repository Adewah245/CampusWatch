package service

import (
	"context"
	"testing"

	"CampusWatch/backend/internal/model"
)

type issueRepoStub struct{ issue *model.Issue }

func (r *issueRepoStub) Create(_ context.Context, value model.Issue) (*model.Issue, error) {
	r.issue = &value
	return &value, nil
}
func (r *issueRepoStub) List(context.Context) ([]model.Issue, error)           { return nil, nil }
func (r *issueRepoStub) GetByID(context.Context, string) (*model.Issue, error) { return r.issue, nil }
func (r *issueRepoStub) Update(context.Context, string, model.UpdateIssueRequest) (*model.Issue, error) {
	return r.issue, nil
}
func (r *issueRepoStub) CreateUpdate(context.Context, string, model.CreateIssueUpdateRequest) (*model.IssueUpdate, error) {
	return &model.IssueUpdate{}, nil
}

func TestIssueServiceCreateDefaultsOpenMedium(t *testing.T) {
	repo := &issueRepoStub{}
	issue, err := NewIssueService(repo).Create(context.Background(), model.CreateIssueRequest{SystemID: "system-1", Title: "High disk usage", Description: "Disk is above threshold"})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}
	if issue.Status != "open" || issue.Priority != "medium" {
		t.Fatalf("unexpected defaults: %+v", issue)
	}
}

func TestIssueServiceRejectsInvalidStatus(t *testing.T) {
	status := "unknown"
	if _, err := NewIssueService(&issueRepoStub{}).Update(context.Background(), "issue-1", model.UpdateIssueRequest{Status: &status}); err == nil {
		t.Fatal("expected invalid status to fail")
	}
}
