//go:build darwin

package session

import "os/exec"

// listLoggedInUsers reports the accounts with an active login session, read
// from the standard who(1) listing, which on macOS reads the same utmpx records
// as Linux.
func listLoggedInUsers() ([]string, error) {
	output, err := exec.Command("who").Output()
	if err != nil {
		return nil, err
	}
	return parseWho(string(output)), nil
}
