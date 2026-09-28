#!/usr/bin/env bash
# =============================================================================
# backup_db.sh
#
# Dumps the e-commerce PostgreSQL database (running in Docker Compose) to a
# timestamped .sql.gz file under ./backups/. Safe to run on a cron schedule.
#
# Usage:
#   ./scripts/backup_db.sh
# =============================================================================

set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_DIR"

# shellcheck disable=SC1091
[ -f .env ] && source .env

POSTGRES_DB="${POSTGRES_DB:-ecommerce}"
POSTGRES_USER="${POSTGRES_USER:-ecommerce_user}"

BACKUP_DIR="${PROJECT_DIR}/backups"
mkdir -p "$BACKUP_DIR"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
OUT_FILE="${BACKUP_DIR}/ecommerce_${TIMESTAMP}.sql.gz"

echo "==> Backing up database '${POSTGRES_DB}' to ${OUT_FILE}"
docker compose exec -T db pg_dump -U "${POSTGRES_USER}" "${POSTGRES_DB}" | gzip > "${OUT_FILE}"

echo "==> Backup complete: ${OUT_FILE}"
echo "    Restore with: gunzip -c ${OUT_FILE} | docker compose exec -T db psql -U ${POSTGRES_USER} ${POSTGRES_DB}"

# Keep only the last 14 backups.
ls -1t "${BACKUP_DIR}"/ecommerce_*.sql.gz 2>/dev/null | tail -n +15 | xargs -r rm --
