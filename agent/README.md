# CampusWatch Monitoring Agent

The CampusWatch Agent runs on each monitored computer. It collects the
operational information the platform is permitted to gather and reports it to
the CampusWatch backend over the agent API.

This is the "eyes and ears" component described in
[`../distributionOfDuties_to_team_member.md`](../distributionOfDuties_to_team_member.md).

---

## What the agent collects

| Area | Data | Source |
| ---- | ---- | ------ |
| Identity | Hostname, operating system, OS version, device type, machine ID | OS published values |
| Health | CPU, memory, disk, battery, network reachability | Aggregate system counters |
| Uptime | Time since boot | Kernel boot records |
| Sessions | Which OS accounts are logged in, and when they change | OS login records |
| Activity | Whether the keyboard/mouse have recently been used | Desktop idle time |
| Local state | A random installation identifier | Generated and stored locally |

## What the agent must never collect

The privacy boundary in [`../README.md`](../README.md) section 6 is a hard
constraint, not a guideline. The agent does **not** read, and this codebase
contains no code that could read:

- Passwords or secret credentials
- Keystrokes
- Private messages
- Browser passwords, cookies or browsing history
- Private files or file contents
- Screen contents or screenshots
- Process command lines or window titles

Only aggregate system counters and the operating system's own login records are
inspected. Every collector lives in `internal/health`, `internal/session` and
`internal/activity`, and each can be read in full in a few minutes.

---

## Building

The agent is a separate Go module with no third-party dependencies, so it builds
into a single static binary with no CGO and nothing to install on the target
machine.

```bash
cd agent
go build ./...
go test ./...

# Build the binary the installers expect.
go build -o campuswatch-agent .
```

Cross-compiling for the supported deployment targets:

```bash
GOOS=linux   GOARCH=amd64 go build -o campuswatch-agent-linux-amd64     .
GOOS=windows GOARCH=amd64 go build -o campuswatch-agent-windows-amd64.exe .
```

> **The `-o campuswatch-agent` is not optional.** The entry point is the
> `main.go` at the module root, so a plain `go build` names the binary after the
> directory and produces `agent`. The installers in `../installer/` look for
> `campuswatch-agent`. If you would rather not pass `-o`, the Linux installer
> accepts `CAMPUSWATCH_AGENT_BINARY` and the Windows installer accepts
> `-Binary` to point at a differently named file.


---

## Configuration

All settings come from the environment. The installers in `../installer/` write
these exact variable names into the agent's environment file, so the names are
part of the deployment contract.

### Required

| Variable | Meaning |
| -------- | ------- |
| `CAMPUSWATCH_SERVER_URL` | Backend base URL, for example `https://campuswatch.example.edu` |
| `CAMPUSWATCH_AGENT_ID` | Agent identifier issued when the agent is approved |
| `CAMPUSWATCH_AGENT_CODE` | Agent code, sent as the `X-Agent-Code` header |
| `CAMPUSWATCH_AGENT_CREDENTIAL` | Per-installation secret, sent as `X-Agent-Credential` |
| `CAMPUSWATCH_SYSTEM_ID` | The system record this agent reports against |

### Optional

| Variable | Default | Meaning |
| -------- | ------- | ------- |
| `CAMPUSWATCH_AGENT_VERSION` | `1.0.0` | Version reported on each heartbeat |
| `CAMPUSWATCH_HEARTBEAT_INTERVAL_SECONDS` | `30` | Heartbeat interval, minimum 5 |
| `CAMPUSWATCH_SESSION_POLL_SECONDS` | `15` | Login/logout and idle poll interval, minimum 5 |
| `CAMPUSWATCH_IDLE_THRESHOLD_SECONDS` | `300` | Inactivity before the session is reported idle |
| `CAMPUSWATCH_STATE_DIR` | Directory of the running binary | Where the installation identifier is stored |

The state directory defaults to the executable's own directory deliberately: the
shipped systemd unit sets `ProtectSystem=strict` with
`ReadWritePaths=$INSTALL_DIR`, so the install directory is the only path the
service may write to.

### Running

```bash
# Directly, with the settings in the environment.
CAMPUSWATCH_SERVER_URL=https://campuswatch.example.edu \
CAMPUSWATCH_AGENT_ID=agent-0001 \
CAMPUSWATCH_AGENT_CODE=LAB-A-PC-001 \
CAMPUSWATCH_AGENT_CREDENTIAL=... \
CAMPUSWATCH_SYSTEM_ID=SYS-000001 \
./campuswatch-agent

# Or pointing at an environment file.
./campuswatch-agent -config /etc/campuswatch-agent.env
```

Values already present in the environment take precedence over the file.

---

## Installation

Use the installers rather than configuring machines by hand:

- Linux: [`../installer/linux/install.sh`](../installer/linux/install.sh)
- Windows: [`../installer/windows/install.ps1`](../installer/windows/install.ps1)

Both install the binary, write the configuration with restrictive permissions,
register a service, and start it. Uninstallers are provided alongside them.

---

## Package layout

| Package | Responsibility |
| ------- | -------------- |
| `main.go` (module root) | Wiring, the heartbeat and session loops, graceful shutdown |
| `internal/config` | Environment loading, validation, defaults |
| `internal/identity` | Hostname, OS, OS version, device type, machine ID |
| `internal/system` | Persisted installation identifier, uptime |
| `internal/health` | CPU, memory, disk, battery, network, thresholds |
| `internal/session` | Login/logout detection and transition tracking |
| `internal/activity` | Active/idle detection and transition tracking |
| `internal/client` | The CampusWatch agent API contract and HTTP client |

Pure parsing lives in untagged `parsers.go` files and pure state machines live in
the tracker types, so the logic most likely to break — kernel output formats and
transition rules — is unit-tested on every development machine rather than only
on the platform it targets.

### API contract

`internal/client/protocol.go` restates the JSON shapes the backend decodes
instead of importing the backend module. This is deliberate:

1. [`../README.md`](../README.md) section 40 requires the teams to integrate
   through contracts, not through each other's implementation.
2. The agent ships to monitored computers. Importing `CampusWatch/backend` would
   drag the entire server source tree into every agent build.

**If the backend contract changes, `protocol.go` must be updated in step.**

---

## Platform support

The agent targets Linux, Windows and macOS. Metrics are read from native
facilities on each: `/proc` and `/sys` on Linux, CIM queries on Windows, and
`sysctl`/IOKit on macOS.

| Platform | Metrics | Sessions | Activity |
| -------- | ------- | -------- | -------- |
| Linux | Native `/proc` and `/sys` reads | `who` | `xprintidle`, falling back to `w` |
| Windows | PowerShell CIM queries | `quser` | `GetLastInputInfo` |
| macOS | `sysctl` and `vm_stat` | `who` | IOKit `HIDIdleTime` |

> **Verification status.** The Linux implementation is the one exercised at
> runtime during development. The Windows and macOS implementations are
> compile-verified (`GOOS=windows go build ./...`, `GOOS=darwin go build ./...`)
> but have not been run on those operating systems. Validate them on real
> hardware before relying on them in production.

### Known limitations

- **Idle detection on Linux is best-effort.** `xprintidle` gives an accurate
  answer on X11 desktops but is an optional utility and does not exist under
  Wayland or on headless machines. The `w` fallback reads terminal session
  activity, which under-reports a user working in graphical applications. Where
  neither source can answer, the agent reports no activity transitions rather
  than inventing them.
- **Unreadable metrics are reported as absent, never guessed.** A metric that
  cannot be read leaves its field at zero, and agent health is reported as
  `degraded` only when neither CPU nor memory could be read.
- **Logins that occur while the agent is stopped are not reported.** The agent
  seeds its session tracker on start rather than announcing everyone already
  logged in as a new login, which would fabricate session history. The backend's
  existing records remain authoritative for that window.
- **Hostname and device type are detected but reported only through the health
  payload's `operating_system` field.** The heartbeat contract has no field for
  the rest. Reporting them would need either an agreed contract extension or the
  administrator-facing system registration endpoint.

---

## Security notes

- The credential is per installation and is never compiled into the binary. It is
  read from the environment and written by the installer with `0600` permissions
  on Linux and an ACL restricted to `SYSTEM` and `Administrators` on Windows.
- The agent warns at start-up when the configured backend URL is not HTTPS.
  [`../README.md`](../README.md) section 37 requires HTTPS in production.
- A credential rejected by the backend (HTTP 401) is surfaced distinctly from a
  transient outage, because retrying rejected credentials never helps.
- The installation identifier is a random locally generated value rather than a
  hardware serial. Hardware identifiers are readable by any process and are
  recycled by some virtualisation platforms; neither is true of a value the
  agent draws itself.
