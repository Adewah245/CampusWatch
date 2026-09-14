package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"CampusWatch/backend/internal/model"
)

type SystemRepository struct{ db *sql.DB }

func NewSystemRepository(db *sql.DB) *SystemRepository { return &SystemRepository{db: db} }

func systemColumns() string {
	return "id, location_id, hostname, device_name, operating_system, os_version, system_status, monitor_status, last_seen_at, created_at, updated_at"
}

func scanSystem(scanner interface{ Scan(...any) error }) (*model.System, error) {
	var value model.System
	err := scanner.Scan(&value.ID, &value.LocationID, &value.Hostname, &value.DeviceName, &value.OperatingSystem, &value.OSVersion, &value.SystemStatus, &value.MonitorStatus, &value.LastSeenAt, &value.CreatedAt, &value.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func (r *SystemRepository) Create(ctx context.Context, value model.System) (*model.System, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO systems (location_id, hostname, device_name, operating_system, os_version, system_status, monitor_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING `+systemColumns(), value.LocationID, value.Hostname, value.DeviceName, value.OperatingSystem, value.OSVersion, value.SystemStatus, value.MonitorStatus)
	created, err := scanSystem(row)
	if err != nil {
		return nil, fmt.Errorf("insert system: %w", err)
	}
	return created, nil
}

func (r *SystemRepository) List(ctx context.Context) ([]model.System, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+systemColumns()+` FROM systems ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query systems: %w", err)
	}
	defer rows.Close()
	var systems []model.System
	for rows.Next() {
		value, err := scanSystem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan system: %w", err)
		}
		systems = append(systems, *value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate systems: %w", err)
	}
	return systems, nil
}

func (r *SystemRepository) GetByID(ctx context.Context, id string) (*model.System, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+systemColumns()+` FROM systems WHERE id = $1`, id)
	value, err := scanSystem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("select system: %w", err)
	}
	return value, nil
}

func (r *SystemRepository) Update(ctx context.Context, id string, input model.UpdateSystemRequest) (*model.System, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE systems SET location_id = COALESCE($1, location_id), hostname = COALESCE($2, hostname),
		device_name = COALESCE($3, device_name), operating_system = COALESCE($4, operating_system),
		os_version = COALESCE($5, os_version), system_status = COALESCE($6, system_status),
		monitor_status = COALESCE($7, monitor_status), updated_at = NOW()
		WHERE id = $8 RETURNING `+systemColumns(), input.LocationID, input.Hostname, input.DeviceName,
		input.OperatingSystem, input.OSVersion, input.SystemStatus, input.MonitorStatus, id)
	value, err := scanSystem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update system: %w", err)
	}
	return value, nil
}

func (r *SystemRepository) Approve(ctx context.Context, id string) (*model.System, error) {
	row := r.db.QueryRowContext(ctx, `UPDATE systems SET system_status = 'active', updated_at = NOW() WHERE id = $1 RETURNING `+systemColumns(), id)
	value, err := scanSystem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("approve system: %w", err)
	}
	return value, nil
}

func (r *SystemRepository) RecordHeartbeat(ctx context.Context, id string, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE systems SET last_seen_at = $1, monitor_status = 'online', updated_at = NOW() WHERE id = $2`, at, id)
	if err != nil {
		return fmt.Errorf("record system heartbeat: %w", err)
	}
	return nil
}

func (r *SystemRepository) SaveHealth(ctx context.Context, report model.HealthReport) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO system_health_snapshots
		(system_id, recorded_at, cpu_percent, memory_percent, disk_percent, battery_percent,
		 network_connected, operating_system, agent_health, uptime_seconds)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, report.SystemID, report.RecordedAt, report.CPUPercent, report.MemoryPercent, report.DiskPercent,
		report.BatteryPercent, report.NetworkConnected, report.OperatingSystem, report.AgentHealth, report.UptimeSeconds)
	if err != nil {
		return fmt.Errorf("save system health: %w", err)
	}
	return nil
}

func (r *SystemRepository) LatestHealth(ctx context.Context, systemID string) (*model.HealthReport, error) {
	var report model.HealthReport
	err := r.db.QueryRowContext(ctx, `
		SELECT system_id, recorded_at, cpu_percent, memory_percent, disk_percent, battery_percent,
		       network_connected, operating_system, agent_health, uptime_seconds, created_at
		FROM system_health_snapshots WHERE system_id = $1 ORDER BY recorded_at DESC LIMIT 1
	`, systemID).Scan(&report.SystemID, &report.RecordedAt, &report.CPUPercent, &report.MemoryPercent,
		&report.DiskPercent, &report.BatteryPercent, &report.NetworkConnected, &report.OperatingSystem,
		&report.AgentHealth, &report.UptimeSeconds, &report.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find latest system health: %w", err)
	}
	return &report, nil
}
