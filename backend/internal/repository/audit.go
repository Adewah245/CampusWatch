package repository

import (
	"context"
	"database/sql"
	"fmt"

	"CampusWatch/backend/internal/model"
)

type AuditRepository struct{ db *sql.DB }

func NewAuditRepository(db *sql.DB) *AuditRepository { return &AuditRepository{db: db} }

func (r *AuditRepository) Record(ctx context.Context, value model.AuditLog) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO audit_logs (user_id, institution_id, method, path) VALUES ($1, $2, $3, $4)`, value.UserID, value.InstitutionID, value.Method, value.Path)
	if err != nil {
		return fmt.Errorf("record audit log: %w", err)
	}
	return nil
}
