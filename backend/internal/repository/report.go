package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"CampusWatch/backend/internal/model"
)

type ReportRepository struct{ db *sql.DB }

func NewReportRepository(db *sql.DB) *ReportRepository { return &ReportRepository{db: db} }

func (r *ReportRepository) Generate(ctx context.Context, kind, period string, from, to time.Time) (*model.Report, error) {
	value := &model.Report{Kind: kind, Period: period, From: from, To: to}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(SUM(duration_seconds), 0) FROM user_sessions WHERE login_at >= $1 AND login_at < $2`, from, to).Scan(&value.UserSessions, &value.UsageSeconds); err != nil {
		return nil, fmt.Errorf("report sessions: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM issues WHERE created_at >= $1 AND created_at < $2`, from, to).Scan(&value.IssuesCreated); err != nil {
		return nil, fmt.Errorf("report issues: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM events WHERE event_type = 'AFTER_HOURS_USAGE' AND occurred_at >= $1 AND occurred_at < $2`, from, to).Scan(&value.AfterHoursEvents); err != nil {
		return nil, fmt.Errorf("report after-hours events: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `
		WITH transitions AS (
			SELECT event_type, occurred_at, LEAD(event_type) OVER (PARTITION BY system_id ORDER BY occurred_at) AS next_type,
			       LEAD(occurred_at) OVER (PARTITION BY system_id ORDER BY occurred_at) AS next_at
			FROM events WHERE event_type IN ('SYSTEM_ONLINE', 'SYSTEM_OFFLINE') AND occurred_at >= $1 AND occurred_at < $2
		)
		SELECT COALESCE(SUM(CASE WHEN event_type = 'SYSTEM_ONLINE' AND next_type = 'SYSTEM_OFFLINE' THEN EXTRACT(EPOCH FROM (next_at - occurred_at))::BIGINT ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN event_type = 'SYSTEM_OFFLINE' AND next_type = 'SYSTEM_ONLINE' THEN EXTRACT(EPOCH FROM (next_at - occurred_at))::BIGINT ELSE 0 END), 0)
		FROM transitions`, from, to).Scan(&value.UptimeSeconds, &value.DowntimeSeconds); err != nil {
		return nil, fmt.Errorf("report uptime: %w", err)
	}
	return value, nil
}
