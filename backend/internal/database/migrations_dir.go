package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveMigrationsDir locates the SQL migrations directory.
//
// The path depends on where the process was started from, and both documented
// invocations are valid: "backend/migrations" exists only when the command runs
// from the repository root, while "migrations" exists only when it runs from
// inside backend/. Hardcoding either one breaks the other, which is how
// `cd backend && go run ./cmd/server` came to fail. Trying both in turn removes
// the need to know which case applies, the same way config.loadDotEnv resolves
// .env.
//
// MIGRATIONS_DIR wins when set, so a deployment that keeps migrations elsewhere
// is unaffected.
func ResolveMigrationsDir() (string, error) {
	if dir := os.Getenv("MIGRATIONS_DIR"); dir != "" {
		return dir, nil
	}

	candidates := []string{
		filepath.Join("backend", "migrations"),
		"migrations",
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
	}

	return "", fmt.Errorf(
		"migrations directory not found; tried %s (set MIGRATIONS_DIR to override, or run from the repository root)",
		strings.Join(candidates, ", "),
	)
}
