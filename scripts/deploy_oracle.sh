#!/usr/bin/env bash
# =============================================================================
# deploy_oracle.sh
#
# Deploys (or updates) the e-commerce app on an Oracle Cloud Always Free VM
# that already hosts another project, WITHOUT touching that project.
#
# This script:
#   - Operates only inside this project's own directory
#   - Builds and (re)starts ONLY this project's Docker Compose stack
#   - Never edits the host's existing Nginx configuration
#   - Never stops any existing Docker container or systemd service
#
# Run it from inside the cloned/uploaded project directory, e.g.:
#   cd /opt/ecommerce-shop
#   ./scripts/deploy_oracle.sh
#
# Prerequisites: Docker + Docker Compose plugin installed, .env configured
# (see .env.example), and scripts/check_ports.sh already run once to
# confirm APP_PORT is free.
# =============================================================================

set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_DIR"

echo "==> Deploying from: $PROJECT_DIR"

if [ ! -f .env ]; then
  echo "ERROR: .env not found. Copy .env.example to .env and configure it first."
  exit 1
fi

# Sanity-check required secrets are not left as placeholders.
if grep -q "replace_with_a_long_random_string" .env; then
  echo "ERROR: JWT_SECRET in .env still has its placeholder value. Set a real secret."
  exit 1
fi
if grep -q "change_me_to_a_strong_password" .env; then
  echo "ERROR: POSTGRES_PASSWORD in .env still has its placeholder value. Set a real password."
  exit 1
fi

echo "==> Confirming APP_PORT is bound to 127.0.0.1 only (never expose the app directly)"
if grep -qE '^APP_BIND_ADDRESS=0\.0\.0\.0' .env; then
  echo "WARNING: APP_BIND_ADDRESS is set to 0.0.0.0 in .env. This will expose the app"
  echo "         port directly to the internet, bypassing Nginx. Recommended: 127.0.0.1."
  read -r -p "Continue anyway? [y/N] " confirm
  [[ "$confirm" =~ ^[Yy]$ ]] || exit 1
fi

echo "==> Pulling/building images"
docker compose build

echo "==> Starting stack (this only affects containers defined in this project's"
echo "    docker-compose.yml -- it does not touch any other project's containers)"
docker compose up -d

APP_PORT_VAL=$(grep -E '^APP_PORT=' .env | cut -d '=' -f2 || echo 8081)
BIND_ADDR=$(grep -E '^APP_BIND_ADDRESS=' .env | cut -d '=' -f2 || echo 127.0.0.1)

echo "==> Waiting for backend health check..."
healthy=false
for i in $(seq 1 30); do
  if curl -sf "http://${BIND_ADDR}:${APP_PORT_VAL}/healthz" > /dev/null 2>&1; then
    echo "==> Backend is healthy at http://${BIND_ADDR}:${APP_PORT_VAL}/healthz"
    healthy=true
    break
  fi
  sleep 2
done

if [ "$healthy" = false ]; then
  echo "WARNING: health check did not succeed within 60s. Check logs:"
  echo "  docker compose logs -f backend"
fi

echo
echo "==> Deployment complete."
echo "    Local check:      curl http://${BIND_ADDR}:${APP_PORT_VAL}/healthz"
echo "    Next step:        add the Nginx server block (see README.md ->"
echo "                      'Oracle Cloud Deployment' step 8) and reload Nginx."
echo "    View logs:        docker compose logs -f"
echo "    Container status: docker compose ps"
