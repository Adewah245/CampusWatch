package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"CampusWatch/backend/internal/model"
)

type AgentRepository struct{ db *sql.DB }

func NewAgentRepository(db *sql.DB) *AgentRepository { return &AgentRepository{db: db} }

const agentColumns = "id, system_id, agent_code, agent_version, status, last_heartbeat_at, created_at, updated_at"

func scanAgent(scanner interface{ Scan(...any) error }) (*model.Agent, error) {
	var value model.Agent
	err := scanner.Scan(&value.ID, &value.SystemID, &value.AgentCode, &value.AgentVersion, &value.Status, &value.LastHeartbeatAt, &value.CreatedAt, &value.UpdatedAt)
	return &value, err
}

func (r *AgentRepository) Create(ctx context.Context, value model.Agent, credentialHash string) (*model.Agent, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO agents (system_id, agent_code, agent_version, status, credential_hash)
		VALUES ($1, $2, $3, 'pending', $4)
		RETURNING `+agentColumns, value.SystemID, value.AgentCode, value.AgentVersion, credentialHash)
	created, err := scanAgent(row)
	if err != nil {
		return nil, fmt.Errorf("insert agent: %w", err)
	}
	return created, nil
}

func (r *AgentRepository) List(ctx context.Context) ([]model.Agent, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+agentColumns+` FROM agents ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query agents: %w", err)
	}
	defer rows.Close()
	var agents []model.Agent
	for rows.Next() {
		value, err := scanAgent(rows)
		if err != nil {
			return nil, fmt.Errorf("scan agent: %w", err)
		}
		agents = append(agents, *value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate agents: %w", err)
	}
	return agents, nil
}

// ListBySystem returns the agents registered against one system, newest first.
//
// The dashboard's system page needs one system's agent, and the only alternative
// — listing every agent and filtering in the browser — would hand a page about a
// single machine the whole fleet's credentials metadata.
func (r *AgentRepository) ListBySystem(ctx context.Context, systemID string) ([]model.Agent, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+agentColumns+` FROM agents WHERE system_id = $1 ORDER BY created_at DESC`, systemID)
	if err != nil {
		return nil, fmt.Errorf("query agents for system: %w", err)
	}
	defer rows.Close()
	var agents []model.Agent
	for rows.Next() {
		value, err := scanAgent(rows)
		if err != nil {
			return nil, fmt.Errorf("scan agent: %w", err)
		}
		agents = append(agents, *value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate agents: %w", err)
	}
	return agents, nil
}

func (r *AgentRepository) Approve(ctx context.Context, id, credentialHash string) (*model.Agent, error) {
	row := r.db.QueryRowContext(ctx, `UPDATE agents SET status = 'approved', credential_hash = $1, updated_at = NOW() WHERE id = $2 RETURNING `+agentColumns, credentialHash, id)
	value, err := scanAgent(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("approve agent: %w", err)
	}
	return value, nil
}

func (r *AgentRepository) FindAuth(ctx context.Context, code string) (*model.AgentAuth, error) {
	var auth model.AgentAuth
	err := r.db.QueryRowContext(ctx, `SELECT `+agentColumns+`, credential_hash FROM agents WHERE agent_code = $1`, code).Scan(
		&auth.Agent.ID, &auth.Agent.SystemID, &auth.Agent.AgentCode, &auth.Agent.AgentVersion,
		&auth.Agent.Status, &auth.Agent.LastHeartbeatAt, &auth.Agent.CreatedAt, &auth.Agent.UpdatedAt,
		&auth.CredentialHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find agent credentials: %w", err)
	}
	return &auth, nil
}

func (r *AgentRepository) RecordHeartbeat(ctx context.Context, id string, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE agents SET last_heartbeat_at = $1, updated_at = NOW() WHERE id = $2`, at, id)
	if err != nil {
		return fmt.Errorf("record agent heartbeat: %w", err)
	}
	return nil
}
