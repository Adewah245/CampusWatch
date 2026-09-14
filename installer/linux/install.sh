#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR="/opt/campuswatch-agent"
CONFIG_FILE="/etc/campuswatch-agent.env"
SERVICE_FILE="/etc/systemd/system/campuswatch-agent.service"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Run this installer as root." >&2
  exit 1
fi

SERVER_URL="${CAMPUSWATCH_SERVER_URL:-}"
AGENT_ID="${CAMPUSWATCH_AGENT_ID:-}"
AGENT_CODE="${CAMPUSWATCH_AGENT_CODE:-}"
AGENT_CREDENTIAL="${CAMPUSWATCH_AGENT_CREDENTIAL:-}"
SYSTEM_ID="${CAMPUSWATCH_SYSTEM_ID:-}"
BINARY="${CAMPUSWATCH_AGENT_BINARY:-./campuswatch-agent}"

usage() {
  echo "Usage: $0 --server-url URL --agent-id ID --agent-code CODE --credential SECRET --system-id ID [--binary PATH]" >&2
  exit 2
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --server-url) SERVER_URL="$2"; shift 2 ;;
    --agent-id) AGENT_ID="$2"; shift 2 ;;
    --agent-code) AGENT_CODE="$2"; shift 2 ;;
    --credential) AGENT_CREDENTIAL="$2"; shift 2 ;;
    --system-id) SYSTEM_ID="$2"; shift 2 ;;
    --binary) BINARY="$2"; shift 2 ;;
    *) usage ;;
  esac
done

[[ -n "$SERVER_URL" && -n "$AGENT_ID" && -n "$AGENT_CODE" && -n "$AGENT_CREDENTIAL" && -n "$SYSTEM_ID" ]] || usage
[[ -f "$BINARY" ]] || { echo "Agent binary not found: $BINARY" >&2; exit 1; }

install -d -m 0755 "$INSTALL_DIR"
install -m 0755 "$BINARY" "$INSTALL_DIR/campuswatch-agent"

umask 077
cat > "$CONFIG_FILE" <<EOF
CAMPUSWATCH_SERVER_URL=$SERVER_URL
CAMPUSWATCH_AGENT_ID=$AGENT_ID
CAMPUSWATCH_AGENT_CODE=$AGENT_CODE
CAMPUSWATCH_AGENT_CREDENTIAL=$AGENT_CREDENTIAL
CAMPUSWATCH_SYSTEM_ID=$SYSTEM_ID
EOF
chmod 0600 "$CONFIG_FILE"

cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=CampusWatch Monitoring Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
EnvironmentFile=$CONFIG_FILE
ExecStart=$INSTALL_DIR/campuswatch-agent
Restart=on-failure
RestartSec=5
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$INSTALL_DIR

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now campuswatch-agent.service
echo "CampusWatch agent installed and started."