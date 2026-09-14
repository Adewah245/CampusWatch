package repository

import (
	"context"
	"database/sql"
	"fmt"

	"CampusWatch/backend/internal/model"
)

type DashboardRepository struct{ db *sql.DB }

func NewDashboardRepository(db *sql.DB) *DashboardRepository { return &DashboardRepository{db: db} }

func (r *DashboardRepository) Summary(ctx context.Context) (*model.DashboardSummary, error) {
	var value model.DashboardSummary
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*),
		       COUNT(*) FILTER (WHERE monitor_status = 'online'),
		       COUNT(*) FILTER (WHERE monitor_status = 'offline'),
		       COUNT(*) FILTER (WHERE monitor_status = 'inactive'),
		       COUNT(*) FILTER (WHERE monitor_status = 'maintenance')
		FROM systems`).Scan(&value.TotalSystems, &value.OnlineSystems, &value.OfflineSystems, &value.InactiveSystems, &value.MaintenanceSystems)
	if err != nil {
		return nil, fmt.Errorf("count systems: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_sessions WHERE state <> 'logged_out'`).Scan(&value.ActiveUsers); err != nil {
		return nil, fmt.Errorf("count active users: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM issues WHERE status IN ('open', 'acknowledged', 'in_progress')`).Scan(&value.OpenIssues); err != nil {
		return nil, fmt.Errorf("count open issues: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM events WHERE event_type = 'AFTER_HOURS_USAGE'`).Scan(&value.AfterHoursUsage); err != nil {
		return nil, fmt.Errorf("count after-hours usage: %w", err)
	}
	return &value, nil
}

func (r *DashboardRepository) Systems(ctx context.Context) ([]model.System, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+systemColumns()+` FROM systems ORDER BY hostname ASC`)
	if err != nil {
		return nil, fmt.Errorf("query dashboard systems: %w", err)
	}
	defer rows.Close()
	var values []model.System
	for rows.Next() {
		value, err := scanSystem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan dashboard system: %w", err)
		}
		values = append(values, *value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dashboard systems: %w", err)
	}
	return values, nil
}
