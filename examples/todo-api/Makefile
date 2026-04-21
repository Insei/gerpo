.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this message
	@awk 'BEGIN {FS = ":.*?## "} \
	     /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## Compile the service binary
	go build -o ./bin/todo-api ./cmd/todo-api

.PHONY: run
run: ## Start the service (expects PG at $$DATABASE_URL or docker-compose)
	go run ./cmd/todo-api

.PHONY: compose-up
compose-up: ## Start Postgres + the service via docker-compose
	docker compose up --build

.PHONY: compose-down
compose-down: ## Stop and remove the docker-compose stack
	docker compose down

# Rebuild bin/custom-gcl only when linter sources or configs change.
GERPOLINT_SRC := $(shell find ../../gerpolint ../../gerpolintplugin -name '*.go' -not -path '*/testdata/*')
CUSTOM_GCL_DEPS := .custom-gcl.yml .golangci.yml ../../go.mod $(GERPOLINT_SRC)

bin/custom-gcl: $(CUSTOM_GCL_DEPS)
	golangci-lint custom

.PHONY: custom-gcl
custom-gcl: bin/custom-gcl ## Build bin/custom-gcl (rebuilt only when the linter or its config changes)

.PHONY: lint
lint: bin/custom-gcl ## Build custom-gcl and run the full linter suite over ./...
	./bin/custom-gcl run ./...
