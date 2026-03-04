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
