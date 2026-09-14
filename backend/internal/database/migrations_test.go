package database

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestApplyMigrationsCreatesTrackingTableAndRunsSQL(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set; skipping migration integration test")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("ping db: %v", err)
	}

	tempDir := t.TempDir()
	migrationPath := filepath.Join(tempDir, "001_institutions.sql")
	if err := os.WriteFile(migrationPath, []byte(`
		CREATE TABLE IF NOT EXISTS institutions (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL
		);
	`), 0644); err != nil {
		t.Fatalf("write migration file: %v", err)
	}

	if err := ApplyMigrations(context.Background(), db, tempDir); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	var count int
	if err := db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'institutions'`).Scan(&count); err != nil {
		t.Fatalf("query institutions table: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected institutions table to exist, got count=%d", count)
	}
}
