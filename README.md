# Gour Accounts

> Centralized identity & authentication platform for the Gour SaaS suite.  
> Works like Google Accounts — one login for all products.

---

## Architecture

```
accounts/
├── backend/          Go (Fiber) REST API
├── frontend/         Vue 3 + Vite SPA
├── db/               PostgreSQL schema
└── Makefile          Dev shortcuts
```

---

## Prerequisites

| Tool | Min version |
|------|-------------|
| Go   | 1.21        |
| Node | 18          |
| PostgreSQL | 14   |

---

## Quick Start

### 1 — Database

```bash
# Create DB and apply schema
make db-setup

# Or manually:
psql -U postgres -c "CREATE DATABASE gour_accounts;"
psql -U postgres -d gour_accounts -f db/schema.sql
```

### 2 — Backend

```bash
cp backend/.env.example backend/.env
# Edit backend/.env with your DB credentials and a strong JWT_SECRET

make install      # go mod tidy + npm install
make backend      # starts on :8080
```

### 3 — Frontend

```bash
make frontend     # starts Vite dev server on :5173
```

### 4 — Run everything

```bash
make dev          # backend + frontend in parallel
```

Open **http://localhost:5173**

---

## API Reference

Base URL: `http://localhost:8080/api/v1`

| Method | Endpoint               | Auth | Description              |
|--------|------------------------|------|--------------------------|
| GET    | `/health`              | —    | Health check             |
| POST   | `/auth/register`       | —    | Create account           |
| POST   | `/auth/login`          | —    | Login, receive JWT       |
| POST   | `/auth/logout`         | —    | Client-side logout       |
| POST   | `/auth/forgot-password`| —    | Request password reset   |
| POST   | `/auth/reset-password` | —    | Reset with token         |
| GET    | `/profile`             | JWT  | Get current user profile |

### Register
```json
POST /api/v1/auth/register
{ "email": "user@example.com", "password": "secret123" }
```

### Login
```json
POST /api/v1/auth/login
{ "email": "user@example.com", "password": "secret123" }
```

### Protected request
```
Authorization: Bearer <jwt_token>
```

---

## Environment Variables (`backend/.env`)

```
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=gour_accounts
JWT_SECRET=change-this-to-a-long-random-secret
# Keep false in production. Products should be created from Admin UI/API.
SEED_DEFAULT_PRODUCTS=false
```

---

## Production Notes (Auth-Service Mode)

- Accounts is identity + access only.
- Products own pricing pages and checkout in Stripe.
- Accounts only tracks which workspace is subscribed to which product/plan.

### Product registry behavior

- Set `SEED_DEFAULT_PRODUCTS=false` (default) so no products are auto-created.
- Create products manually from Admin panel after deploy.

### Stripe metadata contract (required)

When a product creates Stripe checkout/subscription, include:

- `metadata.workspace_id` = Accounts workspace UUID
- `metadata.product_id` = Accounts product UUID
- `metadata.plan_name` = plan label (`starter`, `pro`, etc.)

This is how Accounts maps Stripe subscriptions back to product access.

---

## Roadmap

- [ ] Email delivery for password reset (SendGrid / Resend)
- [ ] OAuth2 social login (Google, GitHub)
- [ ] Workspace & team management
- [ ] Product access control
- [ ] Billing & subscription module
