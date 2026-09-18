//go:build windows

package activity

import (
	"time"

	"CampusWatch/agent/internal/identity"
)

// idleScript calls the Windows GetLastInputInfo API, which returns the tick
// count of the last keyboard or mouse input, and subtracts it from the current
// tick count to produce an idle duration in milliseconds.
//
// The declaration is embedded rather than linked, so the agent stays a single
// static binary with no CGO and no external Go modules. Add-Type compiles the
// small P/Invoke shim on first use.
const idleScript = `Add-Type -Namespace CW -Name Native -MemberDefinition '[DllImport("user32.dll")] public static extern bool GetLastInputInfo(ref LASTINPUTINFO plii); public struct LASTINPUTINFO { public uint cbSize; public uint dwTime; }'
$info = New-Object CW.Native+LASTINPUTINFO
$info.cbSize = [System.Runtime.InteropServices.Marshal]::SizeOf($info)
if (-not [CW.Native]::GetLastInputInfo([ref]$info)) { throw "GetLastInputInfo failed" }
[Environment]::TickCount - $info.dwTime`

// idleDuration reports how long the machine has been idle.
//
// GetLastInputInfo is session-wide, so it measures input anywhere on the
// console rather than input to one application, which is exactly the question
// being asked. If the call cannot be made the duration is reported as
// unavailable and the agent emits no activity transitions.
func idleDuration() (time.Duration, bool) {
	output, err := identity.RunPowerShell(idleScript)
	if err != nil {
		return 0, false
	}
	milliseconds, ok := parseSingleInteger(output)
	if !ok {
		return 0, false
	}
	return time.Duration(milliseconds) * time.Millisecond, true
}
