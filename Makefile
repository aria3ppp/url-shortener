# load env file
ENVFILE ?= .env
include $(ENVFILE)
export

MIGRATE_DSN ?= "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable"
MIGRATE := migrate -path=migrations -database "$(MIGRATE_DSN)"

DOCKER_COMPOSE_SERVICES := docker compose -f docker-compose.services.yml
DOCKER_COMPOSE_SERVER := $(DOCKER_COMPOSE_SERVICES) -f docker-compose.server.yml

.PHONY: default
default: help

# generate help info from comments: thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## help information about make commands
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(firstword $(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: services-ps
services-ps: ## list services containers
	$(DOCKER_COMPOSE_SERVICES) ps

.PHONY: services-up
services-up: ## create and start services
	$(DOCKER_COMPOSE_SERVICES) up -d

.PHONY: services-down
services-down: ## stop and remove services
	$(DOCKER_COMPOSE_SERVICES) down

.PHONY: server-ps
server-ps: ## list server containers
	$(DOCKER_COMPOSE_SERVER) ps

.PHONY: server-up
server-up: ## create and start server
	$(DOCKER_COMPOSE_SERVER) up -d

.PHONY: server-down
server-down: ## stop and remove server
	$(DOCKER_COMPOSE_SERVER) down

.PHONY: server-testdeploy-ps
server-testdeploy-ps: ## list server containers from test deploy
	docker compose -f docker-compose.testdeploy.yml ps

.PHONY: server-testdeploy-up
server-testdeploy-up: ## start server from test deploy
	docker compose -f docker-compose.testdeploy.yml up -d

.PHONY: server-testdeploy-down
server-testdeploy-down: ## stop server from test deploy
	docker compose -f docker-compose.testdeploy.yml down

.PHONY: test
test: ## run tests
	go test -p 1 -covermode=count -coverprofile=coverage.out ./...

.PHONY: test-arg
test-arg: ## run tests by passing $ARG env value to 'go test' command
	go test -covermode=count -coverprofile=coverage.out $(ARG)

.PHONY: test-cover
test-cover: test ## run tests and show test coverage information
	go tool cover -html=coverage.out

.PHONY: test-arg-cover
test-arg-cover: test-arg ## run tests by passing $ARG env value to 'go test' command and show test coverge information
	go tool cover -html=coverage.out

.PHONY: run
run: ## build server and then run entrypoint.sh
	go build -o server .
	./entrypoint.sh

.PHONY: install-dev-deps
install-dev-deps: ## install dev dependencies
	go install github.com/golang/mock/mockgen@v1
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

.PHONY: generate
generate: ## run 'go generate' for all packages
	go generate ./...

.PHONY: migrate
migrate: ## run all new database migrations
	@echo "Running all new database migrations..."
	@$(MIGRATE) up

.PHONY: migrate-down
migrate-down: ## revert database to the last migration step
	@echo "Reverting database to the last migrate step..."
	@$(MIGRATE) down 1

.PHONY: migrate-drop
migrate-drop: ## drop all database migrations
	@echo "dropping database..."
	@$(MIGRATE) drop -f

.PHONY: migrate-new
migrate-new: ## create a new database migration
	@read -p "Enter the name of the new migration: " name; \
	$(MIGRATE) create -ext sql -dir migrations $${name}

.PHONY: migrate-reset
migrate-reset: ## reset database and re-run all migrations
	@echo "Resetting database..."
	@$(MIGRATE) drop -f
	@echo "Running all database migrations..."
	@$(MIGRATE) up