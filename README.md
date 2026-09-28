# ShopSphere E-commerce

A full-stack e-commerce application built with Go, PostgreSQL, Nginx, and plain HTML/CSS/JavaScript. It includes a storefront, shopping cart, checkout flow, JWT-based authentication, role-based admin features, and Docker-based deployment support.

## Overview

This project is designed to run as a small, self-contained stack:

- Go backend API for auth, products, categories, cart, orders, and admin operations
- PostgreSQL database with automatic migrations and optional seeded data
- Nginx as the frontend entry point and reverse proxy to the backend
- Static storefront and admin dashboard in the `frontend/` directory
- Docker Compose setup for local development and simple deployment

## Tech Stack

- Go 1.22+
- PostgreSQL 16
- Nginx
- Docker / Docker Compose
- Gorilla Mux
- JWT + bcrypt
- Vanilla HTML/CSS/JS frontend

## Project Structure

```text
ecommerce/
├── backend/                 Go API source and migrations
│   ├── cmd/api/
│   ├── internal/
│   ├── migrations/
│   ├── seed/
│   ├── tests/
│   ├── Dockerfile
│   └── go.mod
├── frontend/                Static storefront and admin pages
│   ├── css/
│   ├── js/
│   ├── img/
│   ├── index.html
│   ├── product.html
│   ├── cart.html
│   ├── checkout.html
│   ├── login.html
│   ├── register.html
│   ├── orders.html
│   └── admin.html
├── nginx/
│   ├── nginx.local.conf
│   └── shop.example.com.conf.example
├── scripts/
│   ├── backup_db.sh
│   ├── check_ports.sh
│   ├── deploy_oracle.sh
│   ├── rollback_oracle.sh
│   └── ecommerce.service.example
├── docker-compose.yml
├── .env.example
└── README.md
```

## Features

- User registration and login
- JWT-protected authenticated endpoints
- Role-based access for admins and customers
- Product catalog and category browsing
- Search, filtering, sorting, and pagination
- Cart management with stock validation
- Checkout with atomic order creation and stock reduction
- Order history and per-user order visibility
- Admin dashboard for products, categories, orders, users, and stats
- Product image upload support
- Automatic database migration and optional seed script

## Quick Start with Docker

1. Copy the example environment file:

```bash
cp .env.example .env
```

2. Set required values in `.env`, including:

- `JWT_SECRET`
- `POSTGRES_PASSWORD`
- `ADMIN_EMAIL`
- `ADMIN_PASSWORD`

3. Start the stack:

```bash
docker compose up -d --build
```

4. Check the app:

```bash
curl http://localhost:8081/healthz
```

The storefront is served at:

- `http://localhost:8081`

The backend API is available under:

- `http://localhost:8081/api/...`

## Local Development Without Docker

1. Install Go and PostgreSQL.
2. Create a database and user for the app.
3. Configure the backend `.env` file.
4. Run the app:

```bash
cd backend
go run ./cmd/api
```

The Go server can serve the frontend directly when `FRONTEND_DIR` is set, which is useful for local testing without Nginx.

## Environment Variables

Key variables used by the app include:

- `APP_ENV`
- `APP_PORT`
- `APP_BIND_ADDRESS`
- `DATABASE_URL`
- `JWT_SECRET`
- `JWT_EXPIRY_HOURS`
- `ADMIN_EMAIL`
- `ADMIN_PASSWORD`
- `SEED_ON_START`
- `UPLOAD_DIR`
- `MAX_UPLOAD_SIZE_MB`
- `ALLOWED_ORIGINS`

## API Summary

### Public

- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/products`
- `GET /api/products/{slug}`
- `GET /api/categories`
- `GET /api/categories/{slug}`

### Authenticated

- `GET /api/auth/me`
- `GET /api/cart`
- `POST /api/cart/items`
- `PUT /api/cart/items/{productId}`
- `DELETE /api/cart/items/{productId}`
- `DELETE /api/cart`
- `POST /api/orders/checkout`
- `GET /api/orders`
- `GET /api/orders/{id}`

### Admin

- `GET /api/admin/products`
- `POST /api/admin/products`
- `PUT /api/admin/products/{id}`
- `DELETE /api/admin/products/{id}`
- `POST /api/admin/products/{id}/image`
- `GET /api/admin/categories`
- `POST /api/admin/categories`
- `PUT /api/admin/categories/{id}`
- `DELETE /api/admin/categories/{id}`
- `GET /api/admin/orders`
- `GET /api/admin/orders/{id}`
- `PUT /api/admin/orders/{id}/status`
- `GET /api/admin/users`
- `GET /api/admin/stats`

## Database and Migrations

The backend automatically applies migrations when it starts. Sample seed data is also loaded if the app is configured to do so (`SEED_ON_START=true`).

## Deployment Notes

This project includes deployment scripts and examples for a shared Oracle Cloud VM setup. The design keeps the app isolated behind the host Nginx without interfering with any existing project that may already be running on the same machine.

See:

- `nginx/shop.example.com.conf.example` for reverse-proxy configuration examples
- `scripts/` for deployment and rollback helpers

## Useful Commands

```bash
# View logs

docker compose logs -f backend

# Stop the stack

docker compose down

# Remove data volumes

docker compose down -v

# Run backend tests

cd backend
go test ./...
```

## Notes

- The app creates a default admin user automatically on first startup if `ADMIN_EMAIL` and `ADMIN_PASSWORD` are set.
- The upload directory is persisted in Docker volume storage.
- Integration tests against PostgreSQL should be run carefully because they may recreate tables.

## License

This project is intended for local development and self-hosted deployment. Add your preferred license if you plan to distribute or publish it.
