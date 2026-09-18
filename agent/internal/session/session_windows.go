//go:build windows

package session

import (
	"os/exec"
	"strings"
)

// listLoggedInUsers reports the accounts with an interactive session, read from
// the quser utility that ships with Windows.
//
// CombinedOutput is used rather than Output because quser exits with a non-zero
// status when nobody is logged in, writing "No User exists for *" while still
// being a successful answer. Treating that status as a failure would mean the
// agent could never observe the machine becoming idle after the last user logs
// out, so the notice is recognised explicitly and reported as an empty list.
func listLoggedInUsers() ([]string, error) {
	output, err := exec.Command("quser").CombinedOutput()
	text := string(output)

	if err != nil && !isNoSessions(text) {
		return nil, err
	}
	return parseQUser(text), nil
}

// isNoSessions reports whether quser output is the "nobody is logged in" notice
// rather than genuine output.
func isNoSessions(output string) bool {
	lowered := strings.ToLower(output)
	return strings.Contains(lowered, "no user exists") ||
		strings.Contains(lowered, "no users exist")
}
