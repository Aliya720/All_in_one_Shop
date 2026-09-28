#!/usr/bin/env bash
# =============================================================================
# check_ports.sh
#
# Run this on the Oracle Cloud VM BEFORE deploying the e-commerce app, to
# see what's already using ports and Nginx configuration slots, so you
# don't collide with the existing project.
#
# Usage:
#   chmod +x scripts/check_ports.sh
#   ./scripts/check_ports.sh
# =============================================================================

set -euo pipefail

echo "==================================================================="
echo " 1. Listening TCP/UDP ports (look for anything already on 8081)"
echo "==================================================================="
sudo ss -tulpn || true
echo

echo "==================================================================="
echo " 2. Existing Nginx configuration test (must pass BEFORE you touch"
echo "    anything -- if this fails, fix the existing config first, do"
echo "    not proceed with a new site)"
echo "==================================================================="
sudo nginx -t || true
echo

echo "==================================================================="
echo " 3. Enabled Nginx sites (do not remove or edit any of these for"
echo "    your existing project)"
echo "==================================================================="
ls -la /etc/nginx/sites-enabled/ 2>/dev/null || echo "(sites-enabled directory not found -- check /etc/nginx/conf.d/ instead)"
echo
echo "--- /etc/nginx/conf.d/ (some distros use this instead) ---"
ls -la /etc/nginx/conf.d/ 2>/dev/null || true
echo

echo "==================================================================="
echo " 4. Running Docker containers (check for port bindings already"
echo "    in use, e.g. 0.0.0.0:8081->..., before you reuse that port)"
echo "==================================================================="
docker ps 2>/dev/null || echo "(docker not installed or not running as this user)"
echo

echo "==================================================================="
echo " 5. Running systemd services (look for anything you might collide"
echo "    with, e.g. an existing app running as a systemd service rather"
echo "    than in Docker)"
echo "==================================================================="
sudo systemctl --type=service --state=running || true
echo

echo "==================================================================="
echo " Suggested next step:"
echo "   Pick a port from the candidates below that does NOT appear in"
echo "   the 'ss -tulpn' output above, then set APP_PORT to it in your"
echo "   .env file and in nginx/shop.example.com.conf.example."
echo "==================================================================="
for port in 8081 8082 8083 8090 8091 9081; do
  if sudo ss -tulpn 2>/dev/null | grep -q ":$port "; then
    echo "  port $port : IN USE"
  else
    echo "  port $port : available"
  fi
done
