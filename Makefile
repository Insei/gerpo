.DEFAULT_GOAL := help

# Connection string used by integration tests. Override on the command line if
# your local Postgres is on a different port / user.
INTEGRATION_DSN ?= postgres://gerpo:gerpo@localhost:5433/gerpo?sslmode=disable

# Mock-adapter benchmarks that compare a raw executor.Adapter call to gerpo.Repository.
BENCH_PATTERN := ^Benchmark(GetFirst|GetList|Count|Insert|Update|Delete)_(Direct|Gerpo)$$

COMPOSE := docker compose -f tests/integration/docker-compose.yml

.PHONY: help
help: ## Show this message
	@awk 'BEGIN {FS = ":.*?## "} \
	     /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## Compile all packages
	go build -v ./...

.PHONY: test
test: ## Unit tests with the race detector
	go test -race ./...

.PHONY: custom-gcl
custom-gcl: ## Build a golangci-lint binary with the gerpolint plugin embedded (./bin/custom-gcl)
	golangci-lint custom

.PHONY: lint
lint: custom-gcl ## Build custom-gcl and run the full linter suite (incl. gerpolint) over ./...
	./bin/custom-gcl run ./...

.PHONY: lint-gerpolint
lint-gerpolint: ## Run the standalone gerpolint analyzer (no golangci-lint wrapper)
	go run ./cmd/gerpolint ./...

.PHONY: integration-up
integration-up: ## Start the PostgreSQL container used by integration tests
	$(COMPOSE) up -d

.PHONY: integration-down
integration-down: ## Stop the PostgreSQL container
	$(COMPOSE) down

.PHONY: integration
integration: ## Run integration tests (requires integration-up or external PG via INTEGRATION_DSN)
	GERPO_INTEGRATION_DB_URL="$(INTEGRATION_DSN)" go test -tags=integration ./tests/integration/...

.PHONY: integration-full
integration-full: integration-up integration integration-down ## One-shot: bring PG up, run the suite, tear PG down

.PHONY: bench
bench: ## Run Direct vs Gerpo mock benchmarks (5 runs)
	go test -bench='$(BENCH_PATTERN)' -benchmem -run=^$$ -count=5 ./tests/

.PHONY: bench-report
bench-report: ## Print the Direct vs Gerpo summary table (mock adapter, no IO)
	GERPO_BENCH_REPORT=1 go test -run=TestCompareDirectVsGerpo -v ./tests/

BENCH_COMPOSE := docker compose -f tests/bench/docker-compose.yml
BENCH_DSN := postgres://bench:bench@localhost:5434/bench?sslmode=disable

.PHONY: bench-report-pg
bench-report-pg: ## One-shot: spin up an isolated PG, run the pgx v5 vs Gerpo table, tear PG down
	@set -e; \
	echo ">>> starting isolated bench Postgres"; \
	$(BENCH_COMPOSE) up -d --wait; \
	trap '$(BENCH_COMPOSE) down -v --remove-orphans' EXIT; \
	echo ">>> running bench"; \
	GERPO_BENCH_REPORT=1 GERPO_BENCH_PG_DSN="$(BENCH_DSN)" \
	    go test -tags=benchpg -run=TestComparePgxVsGerpo -v ./tests/bench/...

.PHONY: docs-serve
docs-serve: ## Preview the MkDocs site at http://127.0.0.1:8000
	mkdocs serve

.PHONY: docs-build
docs-build: ## Build the MkDocs site with --strict
	mkdocs build --strict

.PHONY: release
release: ## Prepare a release: regenerate CHANGELOG + commit + tag. Usage: make release TAG=vX.Y.Z
	@if [ -z "$(TAG)" ]; then \
		echo "Usage: make release TAG=vX.Y.Z" >&2; \
		exit 2; \
	fi
	./scripts/release.sh $(TAG)
