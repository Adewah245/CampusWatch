package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"CampusWatch/backend/internal/model"
)

type IssueRepository struct{ db *sql.DB }

func NewIssueRepository(db *sql.DB) *IssueRepository { return &IssueRepository{db: db} }

const issueColumns = "id, system_id, title, description, priority, status, reported_by, created_at, updated_at"

func scanIssue(scanner interface{ Scan(...any) error }) (*model.Issue, error) {
	var value model.Issue
	err := scanner.Scan(&value.ID, &value.SystemID, &value.Title, &value.Description, &value.Priority, &value.Status, &value.ReportedBy, &value.CreatedAt, &value.UpdatedAt)
	return &value, err
}

func (r *IssueRepository) Create(ctx context.Context, value model.Issue) (*model.Issue, error) {
	row := r.db.QueryRowContext(ctx, `INSERT INTO issues (system_id, title, description, priority, reported_by) VALUES ($1, $2, $3, $4, $5) RETURNING `+issueColumns, value.SystemID, value.Title, value.Description, value.Priority, value.ReportedBy)
	created, err := scanIssue(row)
	if err != nil {
		return nil, fmt.Errorf("create issue: %w", err)
	}
	return created, nil
}

func (r *IssueRepository) List(ctx context.Context) ([]model.Issue, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+issueColumns+` FROM issues ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query issues: %w", err)
	}
	defer rows.Close()
	var values []model.Issue
	for rows.Next() {
		value, err := scanIssue(rows)
		if err != nil {
			return nil, fmt.Errorf("scan issue: %w", err)
		}
		values = append(values, *value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate issues: %w", err)
	}
	return values, nil
}

func (r *IssueRepository) GetByID(ctx context.Context, id string) (*model.Issue, error) {
	value, err := scanIssue(r.db.QueryRowContext(ctx, `SELECT `+issueColumns+` FROM issues WHERE id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get issue: %w", err)
	}
	return value, nil
}

func (r *IssueRepository) Update(ctx context.Context, id string, input model.UpdateIssueRequest) (*model.Issue, error) {
	value, err := scanIssue(r.db.QueryRowContext(ctx, `UPDATE issues SET title = COALESCE($1, title), description = COALESCE($2, description), priority = COALESCE($3, priority), status = COALESCE($4, status), updated_at = NOW() WHERE id = $5 RETURNING `+issueColumns, input.Title, input.Description, input.Priority, input.Status, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update issue: %w", err)
	}
	return value, nil
}

func (r *IssueRepository) CreateUpdate(ctx context.Context, issueID string, input model.CreateIssueUpdateRequest) (*model.IssueUpdate, error) {
	var value model.IssueUpdate
	err := r.db.QueryRowContext(ctx, `INSERT INTO issue_updates (issue_id, status, comment, updated_by) VALUES ($1, $2, $3, $4) RETURNING id, issue_id, status, comment, updated_by, created_at`, issueID, input.Status, input.Comment, input.UpdatedBy).Scan(&value.ID, &value.IssueID, &value.Status, &value.Comment, &value.UpdatedBy, &value.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create issue update: %w", err)
	}
	if input.Status != nil {
		_, err = r.db.ExecContext(ctx, `UPDATE issues SET status = $1, updated_at = NOW() WHERE id = $2`, *input.Status, issueID)
		if err != nil {
			return nil, fmt.Errorf("update issue status: %w", err)
		}
	}
	return &value, nil
}
