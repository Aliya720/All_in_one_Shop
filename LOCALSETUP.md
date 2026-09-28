# ShopSphere — Go + PostgreSQL E-commerce

A complete e-commerce application: Go REST API, PostgreSQL, a dependency-free HTML/CSS/JS storefront and admin
dashboard, JWT auth, Docker Compose, and Nginx — packaged so it can be deployed to an Oracle Cloud Always Free VM
**that already hosts another project, without touching that project**.

- [Project layout](#project-layout)
- [Features and API summary](#features-and-api-summary)
- [Part 1 — Local deployment](#part-1--local-deployment)
- [Part 2 — Oracle Cloud deployment (shared VM, zero interference)](#part-2--oracle-cloud-deployment-shared-vm-zero-interference)
- [Security notes and known limitations](#security-notes-and-known-limitations)

## Project layout

```text
ecommerce/
├── backend/                 Go API (cmd/api, internal/*), migrations/, seed/, tests/, Dockerfile
├── frontend/                Static storefront + admin dashboard (no build step)
├── database/                Notes on migrations/seed (SQL itself lives in backend/)
├── nginx/
│   ├── nginx.local.conf             Nginx INSIDE the compose stack (serves frontend, proxies /api)
│   └── shop.example.com.conf.example  Template server block for the HOST's Nginx (Oracle)
├── scripts/
│   ├── check_ports.sh               Audit ports / Nginx / containers before deploying
│   ├── deploy_oracle.sh             Build + start ONLY this stack
│   ├── rollback_oracle.sh           Remove ONLY this stack (+ optionally its Nginx site)
│   ├── backup_db.sh                 pg_dump to ./backups
│   └── ecommerce.service.example    Optional systemd wrapper for the compose stack
├── docker-compose.yml
├── .env.example
└── README.md
```

**How the pieces connect.** `docker-compose.yml` runs three containers on a private network: `db` (Postgres),
`backend` (Go, port 8080 internal only) and `nginx` (serves `frontend/`, proxies `/api/` and `/uploads/` to the
backend). The only thing published to the host is the nginx container, at
`${APP_BIND_ADDRESS}:${APP_PORT}` — default `127.0.0.1:8081`. On the Oracle VM, the host's existing Nginx gets one
*additional* server block that forwards `shop.example.com` to that loopback port.

```text
Internet ─► host Nginx (80/443) ─┬─► existing project (unchanged)
                                 └─► shop.example.com ─► 127.0.0.1:8081 ─► [nginx ─► backend ─► db]  (Docker)
```

## Features and API summary

| Area | Capabilities |
|------|--------------|
| Auth | Register, login, `GET /api/auth/me`; JWT (HS256), bcrypt passwords, `customer` / `admin` roles |
| Catalog | Products + categories; search (`search=`), filter (`category=`, `min_price=`, `max_price=`), sort (`sort=price_asc\|price_desc\|name\|newest`), pagination (`page=`, `limit=`) |
| Cart | Per-user cart: add, change quantity, remove, clear (stock is checked on add) |
| Checkout | Single DB transaction: re-validates stock, decrements it, creates order + items (price snapshot), empties cart. Any failure rolls everything back |
| Orders | Order history and detail (users can only read their own); admin can list all and change status |
| Admin | Product/category CRUD, product image upload (jpg/png/webp/gif, size-limited), orders, users, dashboard stats |
| Ops | `GET /healthz`, request logging, panic recovery, CORS, automatic migrations, idempotent seed data |

Main routes: `POST /api/auth/register|login` · `GET /api/products`, `/api/products/{slug}`, `/api/categories` ·
`GET|POST|PUT|DELETE /api/cart…` · `POST /api/orders/checkout`, `GET /api/orders` ·
`/api/admin/{products,categories,orders,users,stats}` (admin only) · `/uploads/*` (product images).

Default admin: created on first start from `ADMIN_EMAIL` / `ADMIN_PASSWORD` in your `.env`.

---

# Part 1 — Local deployment

Pick **Option A (Docker, easiest)** or **Option B (native Go + Postgres)**.

## Option A — Docker Compose

```bash
# 1. Install dependencies: Docker Engine + Compose plugin
#    Linux:  curl -fsSL https://get.docker.com | sudo sh && sudo usermod -aG docker $USER  (re-login afterwards)
#    macOS/Windows: install Docker Desktop
docker --version && docker compose version

# 2-3. Configure .env (the compose stack creates the database for you)
cd ecommerce
cp .env.example .env
sed -i "s|^JWT_SECRET=.*|JWT_SECRET=$(openssl rand -base64 48 | tr -d '\n/+=')|" .env     # macOS: sed -i '' ...
sed -i "s|change_me_to_a_strong_password|$(openssl rand -hex 16)|g" .env                    # DB password
#    Also edit ADMIN_PASSWORD in .env. For sample data set:  SEED_ON_START=true

# 4-6. Migrations run automatically when the backend starts; SEED_ON_START=true loads the sample data.
docker compose up -d --build

# 8. Access
curl http://localhost:8081/healthz          # {"status":"ok"}
# Browser: http://localhost:8081   (port = APP_PORT in .env)
```

Useful: `docker compose logs -f backend` · `docker compose down` (keeps data) · `docker compose down -v` (wipes DB + uploads).

## Option B — Native Go + PostgreSQL

**1. Install dependencies** (Go 1.22+, PostgreSQL 14+):

```bash
# Ubuntu/Debian
sudo apt update && sudo apt install -y golang-go postgresql postgresql-contrib
# macOS
brew install go postgresql@16 && brew services start postgresql@16
go version
```

**2. Create the PostgreSQL database and user:**

```bash
sudo -u postgres psql <<'SQL'
CREATE USER ecommerce_user WITH PASSWORD 'choose_a_password';
CREATE DATABASE ecommerce      OWNER ecommerce_user;
CREATE DATABASE ecommerce_test OWNER ecommerce_user;   -- used by the integration tests
SQL
# macOS (Homebrew): use `psql postgres` instead of `sudo -u postgres psql`
```

**3. Configure `.env`:**

```bash
cd ecommerce/backend
cp .env.example .env
# edit .env: set DATABASE_URL password, JWT_SECRET (openssl rand -base64 48), ADMIN_PASSWORD
# FRONTEND_DIR=../frontend makes the Go server serve the storefront too, so no Nginx is needed locally.
```

**4. Run migrations** — automatic on startup (step 6). To apply them by hand instead:

```bash
psql "postgres://ecommerce_user:choose_a_password@localhost:5432/ecommerce?sslmode=disable" \
     -f migrations/0001_init.up.sql
```

**5. Seed sample data** (5 categories, 12 products; safe to repeat). With `SEED_ON_START=true` / `APP_ENV=development`
it happens on startup; or manually:

```bash
psql "postgres://ecommerce_user:choose_a_password@localhost:5432/ecommerce?sslmode=disable" -f seed/seed.sql
```

**6. Start backend + frontend** (one process):

```bash
go run ./cmd/api
# ... connected to database / applied migration / seed data applied / listening on :8080
```

**7. Run the tests:**

```bash
cd ecommerce/backend
go test ./...                                   # unit tests; DB integration tests are skipped
export TEST_DATABASE_URL="postgres://ecommerce_user:choose_a_password@localhost:5432/ecommerce_test?sslmode=disable"
go test ./... -v                                # includes checkout/transaction tests against Postgres
```

> The integration tests **drop and recreate all tables** in `TEST_DATABASE_URL`. Never point it at real data.

**8. Access locally:** storefront `http://localhost:8080` · admin `http://localhost:8080/admin.html`
(log in with `ADMIN_EMAIL` / `ADMIN_PASSWORD`) · API `http://localhost:8080/api/products`.

> The first `go build` needs internet access for modules. `go.mod` pins `golang.org/x/crypto` to its official GitHub
> mirror via a `replace` line so builds also work behind egress allowlists; you can delete that line if you prefer.

---

# Part 2 — Oracle Cloud deployment (shared VM, zero interference)

**Goal:** add this app to an Always Free VM that already runs another project, with the existing project's ports,
domain, Nginx config and containers left exactly as they are.

**Ground rules built into every step below**

- Nothing here edits, overwrites or replaces an existing Nginx file. We only **add** one new file + one symlink.
- Nginx is only ever **reloaded** (never restarted), and only after `nginx -t` passes.
- No existing container or service is stopped. All our Docker objects are prefixed `ecommerce_`.
- The app port is bound to `127.0.0.1` — never reachable from the internet directly.
- Ports 80/443 are already open for your existing project, so **no new firewall / VCN rules are needed**.

Replace `shop.example.com` with your real subdomain and `/opt/ecommerce-shop` with your preferred directory throughout.

## Step 0 — Inspect the existing server (read-only)

```bash
sudo ss -tulpn                                     # which ports are in use, and by which process
sudo nginx -t                                      # existing config must be VALID before you start
ls -la /etc/nginx/sites-enabled/                   # existing sites (also check /etc/nginx/conf.d/)
docker ps                                          # existing containers and their published ports
sudo systemctl --type=service --state=running      # existing services
```

Or run all of it plus a port check with `./scripts/check_ports.sh` (after step 2). If `nginx -t` already fails,
fix that first — do not proceed.

Also confirm resources (Always Free VMs are small): `free -h && df -h /`. On a 1 GB VM, building the Go image can be
memory-hungry; if the build is killed, add swap first:
`sudo fallocate -l 2G /swapfile && sudo chmod 600 /swapfile && sudo mkswap /swapfile && sudo swapon /swapfile`.

### Pick a free internal port (do NOT assume 8081)

```bash
for p in 8081 8082 8083 8090 8091 9081; do
  sudo ss -tulpn | grep -q ":$p " && echo "$p IN USE" || echo "$p free"
done
```

Choose one that says `free` and use it as `APP_PORT` in step 3 **and** in the Nginx file in step 8. Note that
`ss` shows only *listening* sockets; a port claimed by a stopped container will also appear in `docker ps -a`.
The rest of this guide writes `8081` — substitute your choice.

## Step 1 — Create a dedicated deployment directory

```bash
sudo mkdir -p /opt/ecommerce-shop
sudo chown "$USER":"$USER" /opt/ecommerce-shop
cd /opt/ecommerce-shop
```

Layout on the VM (all separate from the existing project):

| Purpose | Location |
|---------|----------|
| Application code + `docker-compose.yml` | `/opt/ecommerce-shop` |
| Configuration | `/opt/ecommerce-shop/.env` (mode 600) |
| Database data | Docker volume `ecommerce_db_data` |
| Uploaded images | Docker volume `ecommerce_uploads` |
| App logs | `docker compose logs` (rotated: 10 MB × 3 per container) |
| Nginx logs | `/var/log/nginx/shop.example.com.{access,error}.log` |
| DB backups | `/opt/ecommerce-shop/backups` |

## Step 2 — Get the project onto the VM

Docker + Compose plugin must be present (`docker compose version`); if not:
`curl -fsSL https://get.docker.com | sudo sh && sudo usermod -aG docker $USER` and re-login.
Images used (`golang`, `postgres`, `nginx` alpine) are multi-arch, so both AMD and Ampere ARM shapes work.

```bash
# Option 1: upload the ZIP from your computer
scp ecommerce.zip ubuntu@<VM_PUBLIC_IP>:/tmp/
ssh ubuntu@<VM_PUBLIC_IP>
sudo apt-get install -y unzip            # if needed
unzip /tmp/ecommerce.zip -d /tmp/ && cp -a /tmp/ecommerce/. /opt/ecommerce-shop/
# Option 2: git clone <your-repo-url> /opt/ecommerce-shop

cd /opt/ecommerce-shop && chmod +x scripts/*.sh
./scripts/check_ports.sh                 # optional full audit
```

## Step 3 — Configure production `.env`

```bash
cp .env.example .env && chmod 600 .env
sed -i "s|^JWT_SECRET=.*|JWT_SECRET=$(openssl rand -base64 48 | tr -d '\n/+=')|" .env
sed -i "s|change_me_to_a_strong_password|$(openssl rand -hex 24)|g" .env
nano .env
```

Set at least:

```ini
APP_ENV=production
ADMIN_EMAIL=you@yourdomain.com
ADMIN_PASSWORD=<a strong password>
ALLOWED_ORIGINS=https://shop.example.com
APP_BIND_ADDRESS=127.0.0.1        # keep - never 0.0.0.0
APP_PORT=8081                     # the free port you chose
SEED_ON_START=true                # for the FIRST start only if you want sample data
```

`deploy_oracle.sh` refuses to run while placeholder secrets remain.

## Step 4 — Separate PostgreSQL database and user

Nothing to install on the host and nothing shared with the existing project: the stack runs its **own** Postgres
container (`ecommerce_db`) with its own database and user (`POSTGRES_DB` / `POSTGRES_USER` / `POSTGRES_PASSWORD` in
`.env`), reachable only on the private Docker network — it publishes **no** host port, so it cannot clash with any
Postgres already on the VM. Verify after step 5:

```bash
docker compose exec db psql -U ecommerce_user -d ecommerce -c '\dt'
```

<details><summary>Prefer to use an already-installed host PostgreSQL instead?</summary>

Create a **dedicated** role and database (do not reuse the other project's):

```bash
sudo -u postgres psql -c "CREATE USER ecommerce_user WITH PASSWORD 'strong_password';" \
                      -c "CREATE DATABASE ecommerce OWNER ecommerce_user;"
```

Then remove the `db` service and `depends_on` from `docker-compose.yml`, set the backend's
`DATABASE_URL` to `postgres://ecommerce_user:strong_password@host.docker.internal:5432/ecommerce?sslmode=disable`
(add `extra_hosts: ["host.docker.internal:host-gateway"]` to the backend), and allow the Docker bridge subnet in
`pg_hba.conf`. This touches the host's Postgres config, so the bundled container is the safer default.
</details>

## Steps 5–6 — Build, run migrations + seed, start on the internal port

```bash
./scripts/deploy_oracle.sh
```

This builds the images and runs `docker compose up -d`, which starts only the `ecommerce_*` containers. Migrations
run automatically as the backend boots, and — because `SEED_ON_START=true` — the sample data is applied. Then turn
seeding back off (it is idempotent, but sample products are not something you want returning after you delete them):

```bash
sed -i 's/^SEED_ON_START=.*/SEED_ON_START=false/' .env
docker compose up -d                     # recreates only the backend container with the new value
```

Check the startup log for `applied migration: 0001_init.up.sql` and `default admin user created`:

```bash
docker compose ps
docker compose logs backend | tail -20
```

## Step 7 — Verify locally with curl (before touching Nginx)

```bash
curl -s http://127.0.0.1:8081/healthz                       # {"status":"ok"}
curl -s "http://127.0.0.1:8081/api/products?limit=2"        # JSON list (empty items if unseeded)
curl -s -X POST http://127.0.0.1:8081/api/auth/login \
     -H 'Content-Type: application/json' \
     -d '{"email":"you@yourdomain.com","password":"<ADMIN_PASSWORD>"}'   # returns a token
curl -sI http://127.0.0.1:8081/ | head -1                    # HTTP/1.1 200 OK (frontend)

# Confirm it is NOT exposed publicly - must show 127.0.0.1:8081, never 0.0.0.0:8081
sudo ss -tulpn | grep 8081
```

## Steps 8–9 — Add a separate Nginx server block + DNS (existing config untouched)

**DNS first:** in your DNS provider create an `A` record `shop` → the VM's public IP (the same IP the existing site
uses is fine; Nginx tells sites apart by hostname). Check with `dig +short shop.example.com`.

**Back up the current Nginx config** (cheap insurance), then add the new site as a **new file**:

```bash
sudo cp -a /etc/nginx /etc/nginx.backup-$(date +%F)

DOMAIN=shop.yourdomain.com      # your real subdomain
PORT=8081                       # the free port you chose (must equal APP_PORT in .env)

# Create the NEW site file from the template (existing files are never opened for writing)
sed -e "s/shop\.example\.com/$DOMAIN/g" -e "s/127\.0\.0\.1:8081/127.0.0.1:$PORT/" \
    nginx/shop.example.com.conf.example | sudo tee /etc/nginx/sites-available/$DOMAIN > /dev/null
sudo ln -s /etc/nginx/sites-available/$DOMAIN /etc/nginx/sites-enabled/$DOMAIN

sudo nginx -t                              # MUST say "successful"
sudo systemctl reload nginx                # reload, NOT restart: no dropped connections
```

If `nginx -t` fails, **do not reload** — the running Nginx keeps serving your existing site with its old config.
Undo with `sudo rm /etc/nginx/sites-enabled/shop.yourdomain.com` and re-run `sudo nginx -t`.

The generated block is:

```nginx
server {
    listen 80;
    listen [::]:80;
    server_name shop.yourdomain.com;
    access_log /var/log/nginx/shop.yourdomain.com.access.log;
    error_log  /var/log/nginx/shop.yourdomain.com.error.log;
    client_max_body_size 10M;

    location / {
        proxy_pass http://127.0.0.1:8081;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Because the block matches only `server_name shop.yourdomain.com`, requests for your existing domain never reach it.
(If the existing site is the `default_server` for port 80, that stays true.)

```bash
curl -I http://shop.yourdomain.com/healthz        # 200 through Nginx
```

## Step 10 — HTTPS with Let's Encrypt (Certbot)

```bash
certbot --version || sudo apt-get install -y certbot python3-certbot-nginx
sudo certbot --nginx -d shop.yourdomain.com --redirect
```

Naming a single `-d` domain means Certbot edits **only the server block whose `server_name` matches** (ours) and
issues a separate certificate; your existing site's certificate and config are not touched. It also installs a
renewal timer — verify with `sudo certbot renew --dry-run` and `sudo nginx -t`.
Afterwards tighten CORS if you like: `ALLOWED_ORIGINS=https://shop.yourdomain.com` then `docker compose up -d`.

## Step 11 — Automatic restart

Already configured: every container has `restart: unless-stopped`, and Docker's own service starts on boot.

```bash
sudo systemctl is-enabled docker          # should print: enabled   (else: sudo systemctl enable docker)
```

Optional systemd wrapper (only if you want `systemctl status ecommerce`):

```bash
sudo cp scripts/ecommerce.service.example /etc/systemd/system/ecommerce.service
sudo sed -i "s#/opt/ecommerce-shop#$(pwd)#" /etc/systemd/system/ecommerce.service
sudo systemctl daemon-reload && sudo systemctl enable --now ecommerce.service
```

Test resilience (this only kills *our* container): `docker kill ecommerce_backend && sleep 5 && docker ps | grep ecommerce`.

## Step 12 — Verify both applications

```bash
curl -sI https://<your-existing-domain>/ | head -1        # existing project: unchanged, still 200
curl -s  https://shop.yourdomain.com/healthz              # new app: {"status":"ok"}
docker ps                                                 # existing containers still Up + 3 ecommerce_*
sudo systemctl status nginx --no-pager | head -5
sudo ss -tulpn                                            # existing ports unchanged; new one only on 127.0.0.1
```

Then open `https://shop.yourdomain.com`, log in at `/login.html` with the admin credentials, and check `/admin.html`.

## Step 13 — Day-2 commands

```bash
cd /opt/ecommerce-shop

# Logs
docker compose logs -f backend            # app         (also: db, nginx)
docker compose logs --since 1h
sudo tail -f /var/log/nginx/shop.yourdomain.com.error.log

# Restart (only this stack)
docker compose restart backend
docker compose restart                    # all three

# Update to a new version
git pull                                  # or upload/unzip new files over the directory (keep your .env!)
./scripts/backup_db.sh                    # back up first
docker compose build backend && docker compose up -d
# frontend/ is bind-mounted: HTML/CSS/JS changes are live immediately, no rebuild
# new migrations in backend/migrations/ are applied automatically on backend start

# Backup / restore
./scripts/backup_db.sh                    # -> backups/ecommerce_<timestamp>.sql.gz (keeps last 14)
gunzip -c backups/ecommerce_XXXX.sql.gz | docker compose exec -T db psql -U ecommerce_user ecommerce
```

Schedule nightly backups: `crontab -e` → `15 3 * * * cd /opt/ecommerce-shop && ./scripts/backup_db.sh >> backups/cron.log 2>&1`

### Rollback — remove the deployment without affecting the existing project

```bash
cd /opt/ecommerce-shop

# Level 1: stop the app, keep its data (reversible with ./scripts/deploy_oracle.sh)
docker compose down

# Level 2: also remove ONLY our Nginx site (asks for your subdomain, tests config, then reloads)
./scripts/rollback_oracle.sh --remove-nginx

# Level 3: also delete the database + uploads (asks you to type 'yes')
./scripts/rollback_oracle.sh --remove-nginx --purge-data
```

Manual equivalent, if you'd rather not use the script:

```bash
docker compose down                                   # only ecommerce_* containers
docker volume rm ecommerce_db_data ecommerce_uploads  # optional: destroys data
sudo rm /etc/nginx/sites-enabled/shop.yourdomain.com /etc/nginx/sites-available/shop.yourdomain.com
sudo certbot delete --cert-name shop.yourdomain.com   # optional: remove only this certificate
sudo nginx -t && sudo systemctl reload nginx
sudo rm -rf /opt/ecommerce-shop                       # optional
```

If anything ever goes wrong with Nginx itself: `sudo rm /etc/nginx/sites-enabled/shop.yourdomain.com && sudo nginx -t
&& sudo systemctl reload nginx` returns you to exactly the pre-deployment state, and `/etc/nginx.backup-<date>` holds
the full copy from step 8.

### Troubleshooting

| Symptom | Check |
|---------|-------|
| `bind: address already in use` on `docker compose up` | Port taken - pick another (Step 0), update `.env` **and** the Nginx file |
| `502 Bad Gateway` from Nginx | `docker compose ps`; `curl http://127.0.0.1:8081/healthz`; port in Nginx file must equal `APP_PORT` |
| Compose says `JWT_SECRET must be set` | `.env` missing/empty variable (intentional guard) |
| Image build killed / out of memory | Add swap (Step 0) or build on another machine |
| Certbot fails | DNS not pointing at the VM yet, or port 80 blocked by the VCN security list / iptables |
| Uploads rejected | `client_max_body_size` (Nginx, 10M) and `MAX_UPLOAD_SIZE_MB` (app, 5) |

---

## Security notes and known limitations

**Built in:** bcrypt password hashing; JWT verified with algorithm pinning; role checks on every admin route; users can
only read their own orders; parameterized SQL everywhere; request-size and file-type limits on uploads (extension
allow-list, generated filenames); transactional stock handling; unknown JSON fields rejected; panic recovery; the DB
and backend are never published to the host, and the app port is loopback-only.

**Before going live, please note:**

- **No payment gateway.** Checkout creates a `pending` order; an admin marks it `paid`/`shipped`. Integrate Razorpay/Stripe/etc. for real payments.
- **No rate limiting or login lockout.** Consider Nginx `limit_req` on `/api/auth/` (in the new server block only) or fail2ban.
- **JWT is kept in `localStorage`** by the demo frontend (simple, but exposed to XSS). Escaping is applied on rendering, and you may prefer HttpOnly cookies for a hardened build.
- Tokens are not revocable and there is no password reset / email verification.
- Uploaded files are validated by extension, not by content sniffing.
- Change the default admin credentials and use a strong `JWT_SECRET` (the deploy script blocks placeholders).
- The migration runner is forward-only (down files are for manual use).
- Prices are stored as `NUMERIC(12,2)` and shown in ₹ (INR); change `formatPrice` in `frontend/js/api.js` for another currency.
