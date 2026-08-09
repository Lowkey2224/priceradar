# Runs both on the host and inside the toolchain container. On the host every
# recipe is wrapped in `docker compose run`; inside the container (IN_CONTAINER
# is set by Dockerfile.dev) the commands run directly. That way the VS Code
# tasks can call `docker compose run --rm go make <target>` without duplicating
# the command definitions.
ifdef IN_CONTAINER

GO         :=
GO_DB      :=
HOST_ONLY   = $(error `$@` requires Docker and only works on the host)

else

export DOCKER_UID := $(shell id -u)
export DOCKER_GID := $(shell id -g)

COMPOSE    := docker compose --profile tools
GO         := $(COMPOSE) run --rm --no-deps go
GO_DB      := $(COMPOSE) run --rm go
HOST_ONLY   =

endif

.DEFAULT_GOAL := help

.PHONY: help image shell tidy tidy-fix fmt fmt-fix build lint vuln test qa run db-up db-down clean

help:
	@grep -hE '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk -F':.*?## ' '{printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

image: ## (Re)build the toolchain image
	$(HOST_ONLY)$(COMPOSE) build go

shell: ## Interactive shell inside the toolchain container
	$(HOST_ONLY)$(GO_DB) bash

tidy: ## Verify modules and check that go.mod/go.sum are tidy
	$(GO) sh -c 'go mod verify && go mod tidy -diff'

tidy-fix: ## Apply go mod tidy
	$(GO) go mod tidy

fmt: ## Check gofmt formatting
	@$(GO) sh -c 'out="$$(gofmt -l .)"; if [ -n "$$out" ]; then echo "$$out"; exit 1; fi'

fmt-fix: ## Apply gofmt
	$(GO) gofmt -w .

build: ## Compile all packages
	$(GO) go build ./...

lint: ## Run golangci-lint
	$(GO) golangci-lint run ./...

vuln: ## Scan dependencies for known vulnerabilities
	$(GO) govulncheck ./...

test: ## Run tests with race detector and coverage against db-test
	$(GO_DB) gotestsum -- -race -shuffle=on -coverprofile=coverage.out ./...

ifdef IN_CONTAINER
qa: ## Run every check, like the CI pipeline - does not stop at the first failure
	@fail=0; \
	for target in tidy fmt build lint vuln test; do \
		$(MAKE) --no-print-directory $$target || fail=1; \
	done; \
	exit $$fail
else
qa: ## Run every check, like the CI pipeline - does not stop at the first failure
	$(GO_DB) make qa
endif

.env:
	cp .env.dist .env

run: .env ## Run the application against the dev database
	$(GO_DB) go run .

db-up: ## Start the dev database in the background
	$(HOST_ONLY)docker compose up -d db

db-down: ## Stop all containers
	$(HOST_ONLY)$(COMPOSE) down

clean: ## Stop all containers and delete volumes, including dev database and Go caches
	$(HOST_ONLY)$(COMPOSE) down --volumes --remove-orphans
