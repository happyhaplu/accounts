# ── Config ───────────────────────────────────────────────────
GO      := /snap/bin/go
DB_USER := outcraftly
DB_PASS := outcraftly
DB_NAME := accounts
DB_HOST := localhost
DB_PORT := 5432

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	 awk 'BEGIN {FS=":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# ── Dev (run both) ────────────────────────────────────────────
.PHONY: dev
dev: ## Start backend + frontend together
	$(MAKE) -j2 backend frontend

# ── Backend ───────────────────────────────────────────────────
.PHONY: backend
backend: ## Run Go backend (hot: go run)
	-fuser -k 8080/tcp 2>/dev/null || true
	cd backend && $(GO) run main.go

.PHONY: build
build: ## Build Go binary → backend/bin/accounts
	mkdir -p backend/bin
	cd backend && $(GO) build -o bin/accounts .

.PHONY: run
run: build ## Build + run the compiled binary
	./backend/bin/accounts

# ── Frontend ──────────────────────────────────────────────────
.PHONY: frontend
frontend: ## Run Vite dev server
	cd frontend && npm run dev

.PHONY: frontend-build
frontend-build: ## Build frontend for production
	cd frontend && npm run build

# ── Install deps ──────────────────────────────────────────────
.PHONY: install
install: ## Install all dependencies (Go + npm)
	cd backend && $(GO) mod tidy
	cd frontend && npm install

# ── Database ──────────────────────────────────────────────────
.PHONY: db-psql
db-psql: ## Open psql shell for the accounts DB
	PGPASSWORD=$(DB_PASS) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME)

.PHONY: db-status
db-status: ## Show users table schema
	PGPASSWORD=$(DB_PASS) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -c "\d users"

# ── Logs ──────────────────────────────────────────────────────
.PHONY: logs-backend
logs-backend: ## Tail backend log
	tail -f /tmp/backend.log

.PHONY: logs-frontend
logs-frontend: ## Tail frontend log
	tail -f /tmp/frontend.log

# ── Clean ─────────────────────────────────────────────────────
.PHONY: clean
clean: ## Remove build artefacts
	rm -rf backend/bin frontend/dist

# ── Tests ─────────────────────────────────────────────────────
.PHONY: test
test: test-backend test-frontend ## Run all tests (backend + frontend unit)

.PHONY: test-backend
test-backend: ## Run all Go tests (unit + integration)
	cd backend && $(GO) test ./... -count=1 -timeout=120s

.PHONY: test-backend-unit
test-backend-unit: ## Run Go unit tests only (no DB)
	cd backend && $(GO) test ./middleware/... ./handlers/... -v -count=1

.PHONY: test-backend-integration
test-backend-integration: ## Run Go integration tests (in-memory SQLite)
	cd backend && $(GO) test ./integration_test/... -v -count=1 -timeout=120s

.PHONY: test-frontend
test-frontend: ## Run Vitest unit tests
	cd frontend && npm run test

.PHONY: test-frontend-coverage
test-frontend-coverage: ## Run Vitest with coverage report
	cd frontend && npm run test:coverage

.PHONY: test-e2e
test-e2e: ## Run Playwright E2E tests (requires dev server)
	cd frontend && npm run test:e2e

.PHONY: test-regression
test-regression: ## Run backend regression tests only
	cd backend && $(GO) test ./integration_test/... -run Regression -v -count=1 -timeout=120s

.PHONY: test-otp
test-otp: ## Run backend OTP integration tests only
	cd backend && $(GO) test ./integration_test/... -run "(OTP|Verify|Resend|Reset)" -v -count=1 -timeout=120s

.PHONY: test-a11y
test-a11y: ## Run Playwright accessibility tests (requires dev server)
	cd frontend && npx playwright test e2e/accessibility.spec.js

.PHONY: test-load
test-load: ## Run k6 load test (requires k6 installed: https://k6.io)
	@which k6 > /dev/null 2>&1 || (echo "k6 not installed. See https://k6.io/docs/get-started/installation/" && exit 1)
	k6 run scripts/load_test.js

.PHONY: test-load-smoke
test-load-smoke: ## Run a single-VU smoke test with k6
	@which k6 > /dev/null 2>&1 || (echo "k6 not installed. See https://k6.io/docs/get-started/installation/" && exit 1)
	k6 run --vus 1 --iterations 1 scripts/load_test.js

.PHONY: test-all
test-all: test test-e2e test-a11y ## Run every test suite including E2E and accessibility
