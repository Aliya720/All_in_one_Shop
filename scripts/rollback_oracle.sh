#!/usr/bin/env bash
# =============================================================================
# rollback_oracle.sh
#
# Cleanly removes the e-commerce deployment from an Oracle Cloud VM WITHOUT
# affecting any other project on the same host. This:
#   - Stops and removes ONLY this project's Docker containers
#   - Optionally removes this project's Docker volumes (DB data, uploads)
#   - Removes ONLY the Nginx server block for shop.example.com (not any
#     other site's config)
#   - Leaves the existing project's containers, Nginx config, and ports
#     completely untouched
#
# Usage:
#   ./scripts/rollback_oracle.sh              # stop app, keep data volumes
#   ./scripts/rollback_oracle.sh --purge-data # also delete DB + uploads data
#   ./scripts/rollback_oracle.sh --remove-nginx  # also remove the Nginx site
# =============================================================================

set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_DIR"

PURGE_DATA=false
REMOVE_NGINX=false

for arg in "$@"; do
  case "$arg" in
    --purge-data) PURGE_DATA=true ;;
    --remove-nginx) REMOVE_NGINX=true ;;
    *) echo "Unknown option: $arg"; exit 1 ;;
  esac
done

echo "==> Stopping and removing this project's containers only"
echo "    (docker-compose.yml scopes this to the 'ecommerce_*' containers"
echo "    defined here -- no other project's containers are affected)"
docker compose down

if [ "$PURGE_DATA" = true ]; then
  echo "==> Removing this project's Docker volumes (database + uploads)"
  read -r -p "This permanently deletes all order/product/user data. Type 'yes' to confirm: " confirm
  if [ "$confirm" = "yes" ]; then
    docker volume rm ecommerce_db_data ecommerce_uploads 2>/dev/null || true
    echo "    Volumes removed."
  else
    echo "    Skipped -- volumes kept."
  fi
else
  echo "==> Data volumes (ecommerce_db_data, ecommerce_uploads) were kept."
  echo "    Re-running deploy_oracle.sh will restore the same data."
fi

if [ "$REMOVE_NGINX" = true ]; then
  SITE_NAME="shop.example.com"
  echo "==> Removing Nginx site config for ${SITE_NAME} only"
  read -r -p "Enter the subdomain you actually used if different [${SITE_NAME}]: " input_name
  SITE_NAME="${input_name:-$SITE_NAME}"

  if [ -L "/etc/nginx/sites-enabled/${SITE_NAME}" ]; then
    sudo rm "/etc/nginx/sites-enabled/${SITE_NAME}"
    echo "    Removed symlink /etc/nginx/sites-enabled/${SITE_NAME}"
  fi
  if [ -f "/etc/nginx/sites-available/${SITE_NAME}" ]; then
    sudo rm "/etc/nginx/sites-available/${SITE_NAME}"
    echo "    Removed /etc/nginx/sites-available/${SITE_NAME}"
  fi

  echo "==> Testing Nginx config before reload (should only reflect the"
  echo "    REMOVAL of this one site; your other project's config is untouched)"
  sudo nginx -t
  sudo systemctl reload nginx
  echo "    Nginx reloaded. Existing project's site is unaffected."
else
  echo "==> Nginx config was left in place. To remove it manually later:"
  echo "      sudo rm /etc/nginx/sites-enabled/shop.example.com"
  echo "      sudo rm /etc/nginx/sites-available/shop.example.com"
  echo "      sudo nginx -t && sudo systemctl reload nginx"
fi

echo
echo "==> Rollback complete. The existing project on this VM was not modified."
