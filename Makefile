# Ensure GOPATH/bin is in PATH so installed tools are available
GOBIN := $(shell go env GOPATH)/bin
export PATH := $(GOBIN):$(PATH)

MODULE := $(shell go list -m)

NEW_DIRS := internal/user internal/follow internal/topic internal/comment \
            internal/group internal/event internal/chat \
            internal/oauth internal/core internal/platform internal/bootstrap \
            internal/config internal/gates cmd/gates cmd/server

NEW_PKGS := $(addprefix $(MODULE)/, $(NEW_DIRS))

# Tool versions (pinned for deterministic installs)
GOLANGCI_LINT_VERSION := v2.12.2
STATICCHECK_VERSION := v0.7.0
GOIMPORTS_VERSION := v0.46.0
BENCHSTAT_VERSION := latest
GOVULNCHECK_VERSION := v1.4.0
GOFUMPT_VERSION := v0.10.0
GOSEC_VERSION := v2.27.1
GOARCHLINT_VERSION := v1.15.0
LEFTHOOK_VERSION := v2.1.9

# ── Environment ───────────────────────────────────────────────────────

env:
	@echo "=== System ===" && uname -a
	@echo "=== Go ===" && go version && go env
	@echo "=== Module ===" && echo "$(MODULE)"
	@echo "=== Packages ===" && go list ./... | tr '\n' ' ' && echo ""

# ── Tool Installation ─────────────────────────────────────────────────

install: ## Install all dependencies (deterministic, like npm ci)
	@echo "==> Checking prerequisites..."
	@command -v go >/dev/null 2>&1 || { echo "Error: Go not found. Install Go >= 1.25."; exit 1; }
	@GOBIN=$$(go env GOPATH)/bin; \
	echo "$$PATH" | tr ':' '\n' | grep -qxF "$$GOBIN" || { \
		echo "  \342\217\251  $$GOBIN not in PATH. Add to ~/.zshrc:"; \
		echo "     export PATH=\"\$$PATH:\$$(go env GOPATH)/bin\""; \
	}
	@echo "==> Installing Go module dependencies (from go.sum)..."
	go mod download
	@echo "==> Installing root JS tooling (from package-lock.json)..."
	@if [ -f package-lock.json ]; then \
		npm ci; \
	else \
		echo "  [skip] no root package-lock.json found"; \
	fi
	@echo "==> Copying .env.example -> .env (if not exists)..."
	@if [ -f .env.example ]; then \
		cp -n .env.example .env || true; \
	else \
		echo "  [skip] .env.example not found"; \
	fi
	@echo "==> Generating SSL certificates..."
	bash scripts/makecerts.sh
	@echo "==> Installing Go development tools..."
	$(MAKE) tools
	@echo "==> Installing git hooks..."
	$(MAKE) setup-hooks
	@if [ -d frontend-next ] && [ -f frontend-next/package.json ]; then \
		echo "==> Installing frontend-next dependencies..."; \
		command -v bun >/dev/null 2>&1 || { echo "Error: bun not found. Install from https://bun.sh"; exit 1; }; \
		cd frontend-next && bun install; \
	else \
		echo "==> [skip] frontend not scaffolded yet"; \
	fi
	@echo ""
	@echo "Done. Run 'make dev' to start."

setup: tools setup-hooks ## Install Go tools + git hooks only

tools: ## Install pinned Go development tools
	go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	go install honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION)
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)
	go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)
	go install github.com/fe3dback/go-arch-lint@$(GOARCHLINT_VERSION)
	go install golang.org/x/perf/cmd/benchstat@$(BENCHSTAT_VERSION)

setup-hooks: ## Install lefthook pre-commit/pre-push hooks
	go install github.com/evilmartians/lefthook/v2@$(LEFTHOOK_VERSION)
	lefthook install

# ── Docker Compose ────────────────────────────────────────────────────

dev: docker-dev ## Start dev environment (alias to docker-dev)

docker-dev: ## Start dev services with hot-reload
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build

docker-dev-build: ## Rebuild and start dev services
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build --build

docker-down: ## Stop dev services
	docker compose down

docker-clean: ## Remove containers, volumes, local images
	docker compose down -v --rmi local

docker-db: ## Open SQLite inside the running container
	docker exec -it forum-app sqlite3 -line -header db/data/forum.db

# ── Code Formatting ───────────────────────────────────────────────────

format: ## Auto-format all Go files (gofumpt + goimports)
	@goimports -w -local $(MODULE) cmd internal
	@gofumpt -w cmd internal
	@golangci-lint run --fix --timeout=5m || true

check-format: ## Check Go formatting (all code)
	@UNFORMATTED=$$(gofumpt -l cmd internal || true); \
	UNFORMATTED_IMPORTS=$$(goimports -l -local $(MODULE) cmd internal || true); \
	if [ -n "$$UNFORMATTED" ] || [ -n "$$UNFORMATTED_IMPORTS" ]; then \
		[ -n "$$UNFORMATTED" ] && echo "gofumpt errors:" && echo "$$UNFORMATTED"; \
		[ -n "$$UNFORMATTED_IMPORTS" ] && echo "goimports errors:" && echo "$$UNFORMATTED_IMPORTS"; \
		exit 1; \
	fi

check-format-new: ## Check Go formatting (new code only)
	@UNFORMATTED=$$(gofumpt -l $(NEW_DIRS) || true); \
	UNFORMATTED_IMPORTS=$$(goimports -l -local $(MODULE) $(NEW_DIRS) || true); \
	if [ -n "$$UNFORMATTED" ] || [ -n "$$UNFORMATTED_IMPORTS" ]; then \
		[ -n "$$UNFORMATTED" ] && echo "gofumpt errors (new code):" && echo "$$UNFORMATTED"; \
		[ -n "$$UNFORMATTED_IMPORTS" ] && echo "goimports errors (new code):" && echo "$$UNFORMATTED_IMPORTS"; \
		exit 1; \
	fi

# ── Linting & Static Analysis ─────────────────────────────────────────

staticcheck: ## Run staticcheck (all code)
	staticcheck ./...

golangci-lint: ## Run golangci-lint (all code)
	golangci-lint run --timeout=5m

vulncheck: ## Run govulncheck (all code)
	govulncheck ./... || true

gosec: ## Run gosec (all code)
	gosec -quiet ./...

lint: staticcheck golangci-lint vulncheck gosec ## Run all linters (all code)

staticcheck-new: ## Run staticcheck (new code only)
	staticcheck $(NEW_PKGS)

golangci-lint-new: ## Run golangci-lint (new code only)
	golangci-lint run --timeout=5m $(addsuffix /..., $(NEW_DIRS))

vet-new: ## Run go vet (new code only)
	go vet $(NEW_PKGS)

vulncheck-new: ## Run govulncheck (new code only)
	govulncheck $(NEW_PKGS) || true

gosec-new: ## Run gosec (new code only)
	gosec -quiet $(addsuffix /..., $(NEW_DIRS))

lint-new: staticcheck-new golangci-lint-new vet-new vulncheck-new gosec-new ## Run all linters (new code only)

# ── Testing ───────────────────────────────────────────────────────────

test: ## Run backend tests with race detector + coverage (all code)
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out

test-short: ## Run quick/short backend tests (all code)
	go test -short ./...

test-new: ## Run backend tests with race detector (new code only)
	go test -race -coverprofile=coverage.out -covermode=atomic $(NEW_PKGS)
	go tool cover -func=coverage.out

# ── CI Pipeline ───────────────────────────────────────────────────────

ci-mod: ## Verify Go modules are tidy
	go mod tidy
	git diff --exit-code go.mod go.sum || \
		(echo "Error: go.mod/go.sum out of date. Run 'go mod tidy'."; exit 1)

be-ci: ci-mod check-format lint test ## Full backend CI (all code, legacy)

be-ci-new: ci-mod check-format-new lint-new test-new ## Scoped backend CI (new code only)

fe-ci: ## Frontend CI (lint, format:check, typecheck, test)
	@if [ -d frontend-next ] && [ -f frontend-next/package.json ]; then \
		echo "==> Running frontend CI (frontend-next)..."; \
		cd frontend-next && bun run lint && bun run format:check && bun x tsc --noEmit && bun run test; \
	else \
		echo "==> Skipping frontend CI: no frontend scaffolded yet."; \
	fi

ci: be-ci fe-ci ## Full CI (all code)

ci-new: be-ci-new fe-ci ## Scoped CI (new code only)

gates: ## Run all verification gates (go build + Go gates binary)
	go run cmd/gates/main.go --all

check-arch: ## Run go-arch-lint
	go-arch-lint check

# ── Performance & Benchmarking ────────────────────────────────────────

ci-bench: ## Run benchmarks
	go test -run=NONE -bench=. -benchmem ./...

bench-compare: ## Compare benchmarks against main branch
	@git worktree remove -f .git-worktree-main 2>/dev/null || true
	@rm -rf .git-worktree-main
	@git worktree add -d .git-worktree-main main
	@cd .git-worktree-main && go test -run=NONE -bench=. -benchmem -count=5 ./... > ../bench-base.txt
	@git worktree remove -f .git-worktree-main
	@go test -run=NONE -bench=. -benchmem -count=5 ./... > bench-head.txt
	@benchstat bench-base.txt bench-head.txt
	@rm -f bench-base.txt bench-head.txt

bench-profile: ## Generate CPU/mem profiles from benchmarks
	go test -run=NONE -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./...
	@echo "Profiles: cpu.prof mem.prof"

bench-flame: ## Open CPU flame graph
	go tool pprof -http=:8080 cpu.prof

bench-clean: ## Remove benchmark artifacts
	rm -f *.prof bench-*.txt

# ── Build ─────────────────────────────────────────────────────────────

build-backend: ## Build backend binary
	go build -o bin/server cmd/server/main.go

build-frontend: ## Build frontend (Next.js)
	@if [ -f frontend-next/package.json ]; then \
		cd frontend-next && bun run build; \
	else \
		echo "No frontend scaffolded"; \
	fi

build: build-backend build-frontend

# ── Run (Native) ──────────────────────────────────────────────────────

run-backend: ## Run backend natively
	@echo "==> Running backend..."
	go run cmd/server/main.go

run-notifications: ## Run notifications service natively
	@echo "==> Running notifications..."
	cd services/notifications && \
	NOTIFICATIONS_WRITE_TIMEOUT=0 go run cmd/server/main.go

run-broker: ## Start the message broker container
	@echo "📨 Starting broker container on port 5672..."
	@docker rm -f social-network-broker 2>/dev/null || true
	@docker run -d --rm --name social-network-broker -p 5672:5672 danielkotsi/golangmq
	@echo "⏳ Waiting for broker to accept connections..."
	@ok=0; for i in $$(seq 1 60); do \
		if nc -z 127.0.0.1 5672 >/dev/null 2>&1; then ok=1; break; fi; \
		sleep 0.5; \
	done; \
	if [ "$$ok" != "1" ]; then echo "❌ Broker did not become ready"; exit 1; fi
	@echo "✅ Broker ready on port 5672 (PID: $$(docker inspect -f '{{.State.Pid}}' social-network-broker 2>/dev/null || echo 'running'))"

run-frontend: ## Run frontend natively
	@if [ -d frontend-next ] && [ -f frontend-next/package.json ]; then \
		echo "==> Running frontend (Next.js)..."; \
		cd frontend-next && NEXT_PUBLIC_NOTIFICATIONS_ORIGIN=http://localhost:8081 bun run dev; \
	else \
		echo "==> No frontend found"; \
	fi

run: ## Run backend + notifications + frontend concurrently (native)
	@trap 'kill 0' EXIT; \
	$(MAKE) -s run-broker && \
	{ $(MAKE) -s run-notifications & \
	  $(MAKE) -s run-backend & \
	  $(MAKE) -s run-frontend; }

run-all: run ## Alias for run

# ── Database ──────────────────────────────────────────────────────────

db-clean: ## Remove database files
	rm -rf db/data

db-reset: db-clean ## Reset and seed the SQLite database
	@mkdir -p db/data
	$(MAKE) seed

seed: ## Seed database with test data
	sqlite3 db/data/forum.db < db/migrations/schema.sql
	sqlite3 db/data/forum.db < db/migrations/indexes.sql
	sqlite3 db/data/forum.db < db/seeds/dev_data.sql

# ── Cleanup ───────────────────────────────────────────────────────────

clean: ## Remove generated artifacts
	rm -f coverage.out
t:
	@bash -c 'target=$$(make -qp | awk -F: "/^[a-zA-Z0-9][^$#\/\t=]*:([^=]|$$)/ {print $$1}" | sed "s/:$$//" | fzf) && make $$target'

# ── Help ──────────────────────────────────────────────────────────────

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2}'

.PHONY: env dev docker-dev docker-dev-build docker-down docker-clean docker-db \
	install setup tools setup-hooks \
	format check-format check-format-new \
	staticcheck golangci-lint vulncheck gosec lint \
	staticcheck-new golangci-lint-new vet-new vulncheck-new gosec-new lint-new \
	test test-short test-new \
	ci-mod be-ci be-ci-new fe-ci ci ci-new gates check-arch \
	ci-bench bench-compare bench-profile bench-flame bench-clean \
	build-backend build-frontend build \
	run-backend run-notifications run-frontend run run-broker run-all \
	docker-clean docker-db \
	db-clean db-reset seed clean help
