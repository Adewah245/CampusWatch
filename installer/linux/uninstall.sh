#!/usr/bin/env bash
set -euo pipefail

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Run this uninstaller as root." >&2
  exit 1
fi

systemctl disable --now campuswatch-agent.service 2>/dev/null || true
rm -f /etc/systemd/system/campuswatch-agent.service /etc/campuswatch-agent.env
rm -rf /opt/campuswatch-agent
systemctl daemon-reload
echo "CampusWatch agent removed."