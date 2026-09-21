package repository

import (
	"context"
	"database/sql"
	"fmt"

	"CampusWatch/backend/internal/model"
)

type DashboardRepository struct{ db *sql.DB }

func NewDashboardRepository(db *sql.DB) *DashboardRepository { return &DashboardRepository{db: db} }

// OfflineThresholdSeconds is how long a system may go unheard before it counts
// as offline.
//
// Not invented here: README section 13 proposes a 30 second heartbeat with a 2
// minute offline threshold, and the agent implements the heartbeat half of that
// (DefaultHeartbeatInterval in agent/internal/config/config.go). Four missed
// beats is what the documentation calls offline.
const OfflineThresholdSeconds = 120

// summaryQuery counts the fleet from what has actually been observed.
//
// Every figure is derived rather than read from the stored monitor_status,
// because nothing ever transitions a system out of 'online':
// SystemRepository.RecordHeartbeat is the only writer of monitor_status past
// creation and it only ever writes 'online'. A machine that stopped reporting
// therefore stayed online for ever, and 'inactive' was never written at all — so
// the dashboard's Offline and Inactive tiles answered zero no matter what the
// fleet was doing.
//
// The operator's own states still win: maintenance, suspended and retired are
// deliberate and a heartbeat does not override them either way.
//
// `reporting` is the liveness signal. COALESCE matters: last_seen_at is null for
// a system that has never reported, and a null `reporting` would satisfy neither
// the online nor the offline filter, leaving the tiles not summing to the total.
// Such a system counts as offline, which is what a freshly registered one is.
const summaryQuery = `
	WITH fleet AS (
		SELECT
			monitor_status,
			COALESCE(last_seen_at > NOW() - INTERVAL '%d seconds', FALSE) AS reporting,
			EXISTS (
				SELECT 1 FROM user_sessions us
				WHERE us.system_id = s.id AND us.state = 'logged_in_active'
			) AS in_use
		FROM systems s
	)
	SELECT COUNT(*),
	       COUNT(*) FILTER (WHERE monitor_status NOT IN ('maintenance', 'suspended', 'retired') AND reporting),
	       COUNT(*) FILTER (WHERE monitor_status NOT IN ('maintenance', 'suspended', 'retired') AND NOT reporting),
	       COUNT(*) FILTER (WHERE monitor_status NOT IN ('maintenance', 'suspended', 'retired') AND reporting AND NOT in_use),
	       COUNT(*) FILTER (WHERE monitor_status = 'maintenance')
	FROM fleet`

// activeUsersQuery counts people using a machine right now.
//
// Two deliberate narrowings over counting every session that is not logged out.
// 'logged_in_active' excludes `logged_in_idle`, so an unattended machine does not
// count as someone using it — which is what makes the Inactive tile mean
// something. And the join to systems keeps a session on a machine that has
// stopped reporting out of the figure: if the agent is gone we no longer know
// whether anyone is there, and a session that never received its logout would
// otherwise inflate this count for ever.
const activeUsersQuery = `
	SELECT COUNT(*)
	FROM user_sessions us
	JOIN systems s ON s.id = us.system_id
	WHERE us.state = 'logged_in_active'
	  AND s.last_seen_at > NOW() - INTERVAL '%d seconds'`

func (r *DashboardRepository) Summary(ctx context.Context) (*model.DashboardSummary, error) {
	var value model.DashboardSummary
	err := r.db.QueryRowContext(ctx, fmt.Sprintf(summaryQuery, OfflineThresholdSeconds)).Scan(
		&value.TotalSystems, &value.OnlineSystems, &value.OfflineSystems, &value.InactiveSystems, &value.MaintenanceSystems)
	if err != nil {
		return nil, fmt.Errorf("count systems: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, fmt.Sprintf(activeUsersQuery, OfflineThresholdSeconds)).Scan(&value.ActiveUsers); err != nil {
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
