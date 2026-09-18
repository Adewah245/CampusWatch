package session

import (
	"sort"
	"strings"
)

// This file holds the pure parsers for the login listings. They are untagged so
// the formats can be tested on any machine.

// parseWho extracts usernames from the output of who(1), whose lines look like:
//
//	student  tty7     2024-07-02 07:34 (:0)
//	james    pts/0    2024-07-02 09:15 (192.168.1.5)
//
// The username is the first whitespace-separated field. Duplicates are removed,
// because one person may hold several terminal sessions, and the result is
// sorted for deterministic ordering.
func parseWho(output string) []string {
	var usernames []string
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		usernames = append(usernames, fields[0])
	}
	return normalise(usernames)
}

// parseQUser extracts usernames from the output of quser, the Windows
// interactive session listing, whose lines look like:
//
//	 USERNAME              SESSIONNAME        ID  STATE   IDLE TIME   LOGON TIME
//	 student               console             1  Active          .   7/2/2024 7:34 AM
//
// A machine with no interactive sessions reports "No User exists for *".
func parseQUser(output string) []string {
	var usernames []string
	for _, line := range strings.Split(output, "\n") {
		// Older builds prefix the first row with a literal ">".
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), ">"))
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		// Skip the column header and the "no sessions" notice, neither of
		// which is an account name.
		switch strings.ToUpper(fields[0]) {
		case "USERNAME", "NO", "USER":
			continue
		}
		usernames = append(usernames, fields[0])
	}
	return normalise(usernames)
}

// normalise trims, de-duplicates and sorts a username list.
func normalise(usernames []string) []string {
	seen := make(map[string]struct{}, len(usernames))
	unique := make([]string, 0, len(usernames))

	for _, username := range usernames {
		username = strings.TrimSpace(username)
		if username == "" {
			continue
		}
		if _, duplicate := seen[username]; duplicate {
			continue
		}
		seen[username] = struct{}{}
		unique = append(unique, username)
	}

	sort.Strings(unique)
	return unique
}
