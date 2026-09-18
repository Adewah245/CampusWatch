//go:build linux

package session

import "os/exec"

// listLoggedInUsers reports the accounts with an active login session, read
// from the standard who(1) listing.
//
// who reads the system's utmp records, which is the same source that tools such
// as w and last use. No elevated privileges are required and no user content is
// inspected.
func listLoggedInUsers() ([]string, error) {
	output, err := exec.Command("who").Output()
	if err != nil {
		return nil, err
	}
	return parseWho(string(output)), nil
}
