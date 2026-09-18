// Package system provides the state the agent must keep about itself: a stable
// local installation identifier and the machine's uptime.
//
// README section 11 is explicit that CampusWatch must not rest a system's
// identity on values that change, naming IP address, hostname and MAC address
// as examples. The backend issues the authoritative CampusWatch identity; what
// lives here is the agent's own durable anchor, so that an agent restarted on
// the same machine can recognise its previous state even if the machine was
// renamed or re-addressed in the meantime.
package system

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// installationIDFile is the name of the persisted identifier inside StateDir.
const installationIDFile = "installation-id"

// InstallationID returns the stable identifier for this installation,
// generating and persisting one the first time it is called.
//
// The value is a random UUID rather than a hardware serial. Hardware
// identifiers are readable by any process on the machine, so they are weak
// secrets and are recycled by some virtualisation platforms; a random value
// drawn by the agent itself has neither problem.
func InstallationID(stateDir string) (string, error) {
	if stateDir == "" {
		return "", fmt.Errorf("state directory is required")
	}
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		return "", fmt.Errorf("create state directory %s: %w", stateDir, err)
	}

	path := filepath.Join(stateDir, installationIDFile)
	if existing, err := readInstallationID(path); err == nil {
		return existing, nil
	}

	generated, err := newUUID()
	if err != nil {
		return "", err
	}
	if err := writeAtomically(path, generated); err != nil {
		return "", err
	}

	// Read back rather than returning the generated value: if a second agent
	// process started at the same moment, its rename may have won the race, and
	// both processes must agree on the identifier that is actually on disk.
	return readInstallationID(path)
}

// readInstallationID loads a previously persisted identifier.
func readInstallationID(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	id := strings.TrimSpace(string(content))
	if id == "" {
		return "", fmt.Errorf("installation id file %s is empty", path)
	}
	return id, nil
}

// writeAtomically writes value to path via a temporary file and rename, so a
// crash or a concurrent writer can never leave a truncated identifier behind.
func writeAtomically(path, value string) error {
	directory := filepath.Dir(path)

	temp, err := os.CreateTemp(directory, ".installation-id-*")
	if err != nil {
		return fmt.Errorf("create temporary file in %s: %w", directory, err)
	}
	tempName := temp.Name()
	defer os.Remove(tempName) // No-op once the rename below succeeds.

	if err := temp.Chmod(0o600); err != nil {
		temp.Close()
		return fmt.Errorf("secure temporary file: %w", err)
	}
	if _, err := temp.WriteString(value + "\n"); err != nil {
		temp.Close()
		return fmt.Errorf("write installation id: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close installation id file: %w", err)
	}
	if err := os.Rename(tempName, path); err != nil {
		return fmt.Errorf("persist installation id: %w", err)
	}
	return nil
}

// newUUID returns a random RFC 4122 version 4 UUID.
func newUUID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate installation id: %w", err)
	}
	raw[6] = (raw[6] & 0x0f) | 0x40 // Version 4.
	raw[8] = (raw[8] & 0x3f) | 0x80 // RFC 4122 variant.
	return fmt.Sprintf("%x-%x-%x-%x-%x", raw[0:4], raw[4:6], raw[6:8], raw[8:10], raw[10:16]), nil
}
