# Database

PostgreSQL 16 is the only datastore. There is no separate schema tooling to install:

| What | Where |
|------|-------|
| Schema migrations (`*.up.sql`, applied in filename order) | `backend/migrations/` |
| Rollback scripts (`*.down.sql`, run manually - see below) | `backend/migrations/` |
| Sample data (idempotent, `ON CONFLICT DO NOTHING`) | `backend/seed/seed.sql` |

The backend applies any pending `*.up.sql` migration automatically on startup and records
it in the `schema_migrations` table. Migrations are wrapped in a transaction.

## Manual migration / rollback with psql

```bash
# apply (idempotent - the initial migration uses CREATE ... IF NOT EXISTS)
psql "$DATABASE_URL" -f backend/migrations/0001_init.up.sql

# DESTRUCTIVE: drops every application table and all data
psql "$DATABASE_URL" -f backend/migrations/0001_init.down.sql
psql "$DATABASE_URL" -c "DROP TABLE IF EXISTS schema_migrations;"
```

## Adding a migration

Create `backend/migrations/0002_your_change.up.sql` (and a matching `.down.sql`). It is
picked up on the next backend start.
