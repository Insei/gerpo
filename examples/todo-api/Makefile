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

.PHONY: custom-gcl
custom-gcl: ## Build ./bin/custom-gcl with the gerpolint golangci-lint plugin
	golangci-lint custom

.PHONY: lint
lint: custom-gcl ## Build custom-gcl and run the full linter suite over ./...
	./bin/custom-gcl run ./...
